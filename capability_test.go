package jmap

import (
	"encoding/json"
	"testing"
)

func TestGetCapabilities(t *testing.T) {
	type coreCap struct {
		MaxSizeUpload int `json:"maxSizeUpload"`
	}

	tests := []struct {
		name       string
		session    Session
		capability Capability
		want       coreCap
		wantErr    bool
	}{
		{
			name: "present capability",
			session: Session{
				Capabilities: map[Capability]json.RawMessage{
					"urn:ietf:params:jmap:core": json.RawMessage(`{"maxSizeUpload":50000000}`),
				},
			},
			capability: "urn:ietf:params:jmap:core",
			want:       coreCap{MaxSizeUpload: 50000000},
		},
		{
			name: "missing capability",
			session: Session{
				Capabilities: map[Capability]json.RawMessage{},
			},
			capability: "urn:ietf:params:jmap:mail",
			wantErr:    true,
		},
		{
			name: "invalid JSON",
			session: Session{
				Capabilities: map[Capability]json.RawMessage{
					"urn:ietf:params:jmap:core": json.RawMessage(`{invalid`),
				},
			},
			capability: "urn:ietf:params:jmap:core",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCapabilities[coreCap](tt.session, tt.capability)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetCapabilities() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGetAccountCapabilities(t *testing.T) {
	type mailCap struct {
		MaxMailboxDepth int `json:"maxMailboxDepth"`
	}

	tests := []struct {
		name       string
		account    Account
		capability Capability
		want       mailCap
		wantErr    bool
	}{
		{
			name: "present capability",
			account: Account{
				AccountCapabilities: map[Capability]json.RawMessage{
					"urn:ietf:params:jmap:mail": json.RawMessage(`{"maxMailboxDepth":10}`),
				},
			},
			capability: "urn:ietf:params:jmap:mail",
			want:       mailCap{MaxMailboxDepth: 10},
		},
		{
			name: "missing capability",
			account: Account{
				AccountCapabilities: map[Capability]json.RawMessage{},
			},
			capability: "urn:ietf:params:jmap:core",
			wantErr:    true,
		},
		{
			name: "invalid JSON",
			account: Account{
				AccountCapabilities: map[Capability]json.RawMessage{
					"urn:ietf:params:jmap:mail": json.RawMessage(`not-json`),
				},
			},
			capability: "urn:ietf:params:jmap:mail",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAccountCapabilities[mailCap](tt.account, tt.capability)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetAccountCapabilities() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
