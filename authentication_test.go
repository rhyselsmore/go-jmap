package jmap

import (
	"net/http"
	"testing"
)

func TestNewBearerTokenAuthenticator(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"valid token", "my-secret-token", false},
		{"empty token", "", true},
		{"whitespace only", "   ", true},
		{"token with surrounding spaces", " tok ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authn, err := NewBearerTokenAuthenticator(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && authn == nil {
				t.Fatal("expected non-nil authenticator")
			}
		})
	}
}

func TestBearerTokenAuthenticatorAuthenticate(t *testing.T) {
	authn, err := NewBearerTokenAuthenticator("test-token")
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("GET", "https://example.com", nil)
	if err := authn.Authenticate(req); err != nil {
		t.Fatalf("Authenticate() error: %v", err)
	}

	got := req.Header.Get("Authorization")
	want := "Bearer test-token"
	if got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestWithBearerTokenAuthentication(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"valid token", "tok", false},
		{"empty token", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithBearerTokenAuthentication(tt.token)
			cl := &Client{}
			err := opt(cl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
