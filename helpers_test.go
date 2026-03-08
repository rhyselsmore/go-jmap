package jmap

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
)

// testInvocation is a mock Invocation for use in tests.
type testInvocation struct {
	name      string
	id        string
	response  json.RawMessage
	decodeErr error
}

func (t *testInvocation) Name() string { return t.name }
func (t *testInvocation) ID() string   { return t.id }
func (t *testInvocation) DecodeResponse(b json.RawMessage) error {
	t.response = b
	return t.decodeErr
}

// testResolver is a mock Resolver that returns a fixed URL.
type testResolver struct {
	url *url.URL
}

func (r *testResolver) Resolve(_ context.Context) (*url.URL, error) {
	return r.url, nil
}

// failResolver is a Resolver that always returns an error.
type failResolver struct{}

func (failResolver) Resolve(_ context.Context) (*url.URL, error) {
	return nil, errors.New("resolve failed")
}

// countingResolver counts how many times Resolve is called.
type countingResolver struct {
	calls *int
}

func (r *countingResolver) Resolve(_ context.Context) (*url.URL, error) {
	*r.calls++
	u, _ := url.Parse("https://example.com/.well-known/jmap")
	return u, nil
}
