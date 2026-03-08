package jmap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// failAuthenticator is an Authenticator that always returns an error.
type failAuthenticator struct{}

func (failAuthenticator) Authenticate(_ *http.Request) error {
	return errors.New("auth failed")
}

// unmarshalableInvocation is an Invocation that fails to marshal.
type unmarshalableInvocation struct {
	Bad chan int `json:"bad"`
}

func (u *unmarshalableInvocation) Name() string                       { return "Bad/method" }
func (u *unmarshalableInvocation) ID() string                         { return "bad" }
func (u *unmarshalableInvocation) DecodeResponse(_ json.RawMessage) error { return nil }

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    []ClientOpt
		wantErr bool
	}{
		{
			name:    "missing resolver",
			opts:    []ClientOpt{WithBearerTokenAuthentication("tok")},
			wantErr: true,
		},
		{
			name:    "missing authenticator",
			opts:    []ClientOpt{WithStaticResolver("https://example.com")},
			wantErr: true,
		},
		{
			name: "valid minimal config",
			opts: []ClientOpt{
				WithStaticResolver("https://example.com"),
				WithBearerTokenAuthentication("tok"),
			},
		},
		{
			name: "with custom HTTP client",
			opts: []ClientOpt{
				WithStaticResolver("https://example.com"),
				WithBearerTokenAuthentication("tok"),
				WithHTTPClient(&http.Client{}),
			},
		},
		{
			name:    "option error propagates",
			opts:    []ClientOpt{WithStaticResolver("")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, err := NewClient(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && cl == nil {
				t.Fatal("expected non-nil client")
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 42}
	opt := WithHTTPClient(custom)
	cl := &Client{}
	if err := opt(cl); err != nil {
		t.Fatalf("error: %v", err)
	}
	if cl.http == custom {
		t.Error("should clone the client, not share the pointer")
	}
	if cl.http.Timeout != 42 {
		t.Error("cloned client should preserve Timeout")
	}
}

func TestClientNewRequest(t *testing.T) {
	cl := &Client{}

	t.Run("with context", func(t *testing.T) {
		req, err := cl.newRequest(context.Background(), "GET", "https://example.com", nil)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if req.Method != "GET" {
			t.Errorf("Method = %q, want GET", req.Method)
		}
	})

	t.Run("nil context falls back", func(t *testing.T) {
		req, err := cl.newRequest(nil, "POST", "https://example.com", nil)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if req.Context() == nil {
			t.Error("expected non-nil context")
		}
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		_, err := cl.newRequest(context.Background(), "GET", "://bad", nil)
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestClientAuthenticationTransport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cl, err := NewClient(
		WithStaticResolver("https://example.com"),
		WithBearerTokenAuthentication("test-token"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	req, _ := http.NewRequest("GET", srv.URL, nil)
	resp, err := cl.http.Do(req)
	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (auth should have been injected)", resp.StatusCode)
	}
}

func TestClientDo(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/session":
			json.NewEncoder(w).Encode(map[string]any{
				"capabilities":    map[string]any{},
				"accounts":        map[string]any{},
				"primaryAccounts": map[string]any{},
				"username":        "test@example.com",
				"apiUrl":          srv.URL + "/api",
				"downloadUrl":     "",
				"uploadUrl":       "",
				"eventSourceUrl":  "",
				"state":           "s1",
			})
		case "/api":
			w.Write([]byte(`{
				"methodResponses": [["Mailbox/get", {"list": []}, "c1"]],
				"sessionState": "s1"
			}`))
		}
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")

	cl, err := NewClient(
		func(c *Client) error {
			c.resolver = &testResolver{url: u}
			return nil
		},
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	inv := &testInvocation{name: "Mailbox/get"}
	req := NewRequest("urn:ietf:params:jmap:core")
	req.Add(inv)

	_, err = cl.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if inv.response == nil {
		t.Error("invocation response should have been decoded")
	}
}

func TestClientDoNon2xx(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/session":
			json.NewEncoder(w).Encode(map[string]any{
				"capabilities":    map[string]any{},
				"accounts":        map[string]any{},
				"primaryAccounts": map[string]any{},
				"username":        "test@example.com",
				"apiUrl":          srv.URL + "/api",
				"state":           "s1",
			})
		case "/api":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error {
			c.resolver = &testResolver{url: u}
			return nil
		},
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	req := NewRequest("urn:ietf:params:jmap:core")
	req.Add(&testInvocation{name: "Mailbox/get"})

	_, err = cl.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Do() expected error for non-2xx status")
	}
}

func TestClientGetSessionNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error {
			c.resolver = &testResolver{url: u}
			return nil
		},
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	_, err = cl.GetSession(context.Background())
	if err == nil {
		t.Fatal("GetSession() expected error for non-2xx")
	}
}

func TestClientConfigureDefaultTransport(t *testing.T) {
	customTransport := &http.Transport{}
	cl, err := NewClient(
		WithStaticResolver("https://example.com"),
		WithBearerTokenAuthentication("tok"),
		WithHTTPClient(&http.Client{Transport: customTransport}),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if cl.http.Transport == nil {
		t.Error("transport should be wrapped, not nil")
	}
}

func TestWithAuthenticator(t *testing.T) {
	authn, _ := NewBearerTokenAuthenticator("tok")
	opt := WithAuthenticator(authn)
	cl := &Client{}
	if err := opt(cl); err != nil {
		t.Fatalf("error: %v", err)
	}
	if cl.authn != authn {
		t.Error("authenticator not set")
	}
}

func TestClientAuthenticatorError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cl, err := NewClient(
		WithStaticResolver("https://example.com"),
		WithAuthenticator(failAuthenticator{}),
	)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}

	req, _ := http.NewRequest("GET", srv.URL, nil)
	_, err = cl.http.Do(req)
	if err == nil {
		t.Fatal("expected error from failing authenticator")
	}
}

func TestClientDoMarshalError(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"capabilities": map[string]any{}, "accounts": map[string]any{},
			"primaryAccounts": map[string]any{}, "username": "u",
			"apiUrl": srv.URL + "/api", "state": "s1",
		})
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error { c.resolver = &testResolver{url: u}; return nil },
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest()
	req.Add(&unmarshalableInvocation{Bad: make(chan int)})

	_, err = cl.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Do() expected error for unmarshalable request")
	}
}

func TestClientDoInvalidResponseJSON(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/session":
			json.NewEncoder(w).Encode(map[string]any{
				"capabilities": map[string]any{}, "accounts": map[string]any{},
				"primaryAccounts": map[string]any{}, "username": "u",
				"apiUrl": srv.URL + "/api", "state": "s1",
			})
		case "/api":
			w.Write([]byte(`{not valid json`))
		}
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error { c.resolver = &testResolver{url: u}; return nil },
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest()
	req.Add(&testInvocation{name: "Mailbox/get"})

	_, err = cl.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Do() expected error for invalid response JSON")
	}
}

func TestClientDoCorrelateError(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/session":
			json.NewEncoder(w).Encode(map[string]any{
				"capabilities": map[string]any{}, "accounts": map[string]any{},
				"primaryAccounts": map[string]any{}, "username": "u",
				"apiUrl": srv.URL + "/api", "state": "s1",
			})
		case "/api":
			w.Write([]byte(`{
				"methodResponses": [["error", {"type":"serverFail"}, "c1"]],
				"sessionState": "s1"
			}`))
		}
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error { c.resolver = &testResolver{url: u}; return nil },
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest()
	req.Add(&testInvocation{name: "Mailbox/get"})

	_, err = cl.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Do() expected error for invocation error")
	}
}

func TestClientGetSessionInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{not json`))
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error { c.resolver = &testResolver{url: u}; return nil },
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cl.GetSession(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid session JSON")
	}
}

func TestClientResolveSessionURLError(t *testing.T) {
	cl, err := NewClient(
		func(c *Client) error {
			c.resolver = &failResolver{}
			return nil
		},
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cl.GetSession(context.Background())
	if err == nil {
		t.Fatal("expected error from failing resolver")
	}
}

func TestClientDoHTTPError(t *testing.T) {
	// Server that serves a session pointing to an unreachable API URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"capabilities": map[string]any{}, "accounts": map[string]any{},
			"primaryAccounts": map[string]any{}, "username": "u",
			"apiUrl": "http://127.0.0.1:1/api", "state": "s1",
		})
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL + "/session")
	cl, err := NewClient(
		func(c *Client) error { c.resolver = &testResolver{url: u}; return nil },
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	req := NewRequest()
	req.Add(&testInvocation{name: "Mailbox/get"})

	_, err = cl.Do(context.Background(), req)
	if err == nil {
		t.Fatal("Do() expected error for unreachable API")
	}
}

func TestClientResolveSessionURLCaches(t *testing.T) {
	calls := 0
	cl, err := NewClient(
		func(c *Client) error {
			c.resolver = &countingResolver{calls: &calls}
			return nil
		},
		WithBearerTokenAuthentication("tok"),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Call resolveSessionURL twice to exercise the cache hit path.
	u1, err := cl.resolveSessionURL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	u2, err := cl.resolveSessionURL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if u1.String() != u2.String() {
		t.Error("cached URL should be the same")
	}
	if calls != 1 {
		t.Errorf("resolver called %d times, want 1", calls)
	}
}
