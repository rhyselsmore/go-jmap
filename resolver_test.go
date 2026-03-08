package jmap

import (
	"context"
	"testing"
)

func TestNewStaticResolver(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantURL string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://api.fastmail.com",
			wantURL: "https://api.fastmail.com/.well-known/jmap",
		},
		{
			name:    "valid HTTPS URL with port",
			url:     "https://api.fastmail.com:443",
			wantURL: "https://api.fastmail.com:443/.well-known/jmap",
		},
		{
			name:    "trailing slash accepted",
			url:     "https://api.fastmail.com/",
			wantURL: "https://api.fastmail.com/.well-known/jmap",
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "HTTP scheme rejected",
			url:     "http://api.fastmail.com",
			wantErr: true,
		},
		{
			name:    "path rejected",
			url:     "https://api.fastmail.com/some/path",
			wantErr: true,
		},
		{
			name:    "query parameters rejected",
			url:     "https://api.fastmail.com?key=val",
			wantErr: true,
		},
		{
			name:    "fragment rejected",
			url:     "https://api.fastmail.com#section",
			wantErr: true,
		},
		{
			name:    "missing host",
			url:     "https://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewStaticResolver(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			got, err := r.Resolve(context.Background())
			if err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}
			if got.String() != tt.wantURL {
				t.Errorf("Resolve() = %q, want %q", got.String(), tt.wantURL)
			}
		})
	}
}

func TestWithStaticResolver(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid URL", "https://api.example.com", false},
		{"empty URL", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithStaticResolver(tt.url)
			cl := &Client{}
			err := opt(cl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
