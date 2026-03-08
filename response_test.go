package jmap

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestInvocationError(t *testing.T) {
	tests := []struct {
		name string
		err  InvocationError
		want string
	}{
		{
			name: "with detail",
			err:  InvocationError{CallID: "c1", Type: "invalidArguments", Detail: "unknown property"},
			want: `jmap: invocation "c1" error: invalidArguments: unknown property`,
		},
		{
			name: "without detail",
			err:  InvocationError{CallID: "c1", Type: "serverFail"},
			want: `jmap: invocation "c1" error: serverFail`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResponseUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantN   int
		wantErr bool
	}{
		{
			name: "valid response with two methods",
			input: `{
				"methodResponses": [
					["Mailbox/get", {"list": []}, "c1"],
					["Email/query", {"ids": ["e1"]}, "c2"]
				],
				"sessionState": "abc123"
			}`,
			wantN: 2,
		},
		{
			name:    "invalid JSON",
			input:   `{not json}`,
			wantErr: true,
		},
		{
			name: "wrong element count",
			input: `{
				"methodResponses": [["Mailbox/get", {"list": []}]],
				"sessionState": "s1"
			}`,
			wantErr: true,
		},
		{
			name: "bad method name type",
			input: `{
				"methodResponses": [[123, {"list": []}, "c1"]],
				"sessionState": "s1"
			}`,
			wantErr: true,
		},
		{
			name: "bad call ID type",
			input: `{
				"methodResponses": [["Mailbox/get", {"list": []}, 123]],
				"sessionState": "s1"
			}`,
			wantErr: true,
		},
		{
			name: "empty method responses",
			input: `{
				"methodResponses": [],
				"sessionState": "s1"
			}`,
			wantN: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp Response
			err := json.Unmarshal([]byte(tt.input), &resp)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(resp.MethodResponses) != tt.wantN {
				t.Errorf("len(MethodResponses) = %d, want %d", len(resp.MethodResponses), tt.wantN)
			}
		})
	}
}

func TestResponseUnmarshalJSONFields(t *testing.T) {
	input := `{
		"methodResponses": [["Mailbox/get", {"list": []}, "c1"]],
		"sessionState": "state42"
	}`

	var resp Response
	if err := json.Unmarshal([]byte(input), &resp); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}

	if resp.SessionState != "state42" {
		t.Errorf("SessionState = %q, want %q", resp.SessionState, "state42")
	}

	mr := resp.MethodResponses[0]
	if mr.Name != "Mailbox/get" {
		t.Errorf("Name = %q, want %q", mr.Name, "Mailbox/get")
	}
	if mr.ID != "c1" {
		t.Errorf("ID = %q, want %q", mr.ID, "c1")
	}
	// Verify Args contains valid JSON.
	var v any
	if err := json.Unmarshal(mr.Args, &v); err != nil {
		t.Errorf("Args not valid JSON: %v", err)
	}
}

func TestCorrelate(t *testing.T) {
	t.Run("successful decode", func(t *testing.T) {
		inv := &testInvocation{name: "Mailbox/get", id: "c1"}
		req := NewRequest()
		req.Add(inv)

		resp := &Response{
			MethodResponses: []MethodResponse{
				{Name: "Mailbox/get", Args: json.RawMessage(`{"list":[]}`), ID: "c1"},
			},
		}
		if err := resp.correlate(req); err != nil {
			t.Fatalf("correlate() error: %v", err)
		}
		if inv.response == nil {
			t.Error("invocation response was not set")
		}
	})

	t.Run("invocation error", func(t *testing.T) {
		inv := &testInvocation{name: "Mailbox/get", id: "c1"}
		req := NewRequest()
		req.Add(inv)

		resp := &Response{
			MethodResponses: []MethodResponse{
				{Name: "error", Args: json.RawMessage(`{"type":"serverFail","description":"oops"}`), ID: "c1"},
			},
		}
		err := resp.correlate(req)
		if err == nil {
			t.Fatal("correlate() expected error")
		}
		var invErr *InvocationError
		if !errors.As(err, &invErr) {
			t.Fatalf("expected InvocationError, got %T", err)
		}
		if invErr.Type != "serverFail" {
			t.Errorf("Type = %q, want %q", invErr.Type, "serverFail")
		}
		if invErr.Detail != "oops" {
			t.Errorf("Detail = %q, want %q", invErr.Detail, "oops")
		}
	})

	t.Run("unknown call ID skipped", func(t *testing.T) {
		req := NewRequest()
		resp := &Response{
			MethodResponses: []MethodResponse{
				{Name: "Mailbox/get", Args: json.RawMessage(`{}`), ID: "unknown"},
			},
		}
		if err := resp.correlate(req); err != nil {
			t.Fatalf("should not error for unknown IDs: %v", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		inv := &testInvocation{
			name:      "Mailbox/get",
			id:        "c1",
			decodeErr: errors.New("decode failed"),
		}
		req := NewRequest()
		req.Add(inv)

		resp := &Response{
			MethodResponses: []MethodResponse{
				{Name: "Mailbox/get", Args: json.RawMessage(`{}`), ID: "c1"},
			},
		}
		err := resp.correlate(req)
		if err == nil {
			t.Fatal("correlate() expected error")
		}
		if !strings.Contains(err.Error(), "decode failed") {
			t.Errorf("error = %q, want to contain %q", err.Error(), "decode failed")
		}
	})

	t.Run("invalid error JSON", func(t *testing.T) {
		inv := &testInvocation{name: "Mailbox/get", id: "c1"}
		req := NewRequest()
		req.Add(inv)

		resp := &Response{
			MethodResponses: []MethodResponse{
				{Name: "error", Args: json.RawMessage(`{invalid}`), ID: "c1"},
			},
		}
		err := resp.correlate(req)
		if err == nil {
			t.Fatal("expected error for invalid error JSON")
		}
	})

	t.Run("no errors returns nil", func(t *testing.T) {
		req := NewRequest()
		resp := &Response{}
		if err := resp.correlate(req); err != nil {
			t.Fatalf("correlate() error: %v", err)
		}
	})
}
