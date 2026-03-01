package jmap

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Resolver resolves the JMAP session URL for a given server.
// Implementations may perform DNS lookups, SRV record queries, or simply
// return a statically configured URL.
type Resolver interface {
	Resolve(ctx context.Context) (*url.URL, error)
}

// wellKnownPath is the IANA-registered well-known URI for JMAP session
// discovery, as defined in RFC 8620 Section 2.2.
const wellKnownPath = "/.well-known/jmap"

// WithStaticResolver is a ClientOpt that configures the client to use a
// StaticResolver with the given HTTPS host URL.
func WithStaticResolver(raw string) ClientOpt {
	return func(c *Client) error {
		var err error
		c.resolver, err = NewStaticResolver(raw)
		return err
	}
}

// NewStaticResolver creates a Resolver that always returns the JMAP
// well-known URL for the given host. The input must be an HTTPS URL
// containing only a scheme and host (with optional port). Paths, query
// parameters, and fragments are not permitted.
//
// Example:
//
//	r, err := NewStaticResolver("https://api.fastmail.com")
func NewStaticResolver(raw string) (*StaticResolver, error) {
	if raw == "" {
		return nil, errors.New("jmap: resolver URL must not be empty")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("jmap: invalid resolver URL: %w", err)
	}

	if parsed.Scheme != "https" {
		return nil, errors.New("jmap: resolver URL must use https scheme")
	}

	if parsed.Host == "" {
		return nil, errors.New("jmap: resolver URL must include a host")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return nil, errors.New("jmap: resolver URL must not include a path")
	}

	if parsed.RawQuery != "" {
		return nil, errors.New("jmap: resolver URL must not include query parameters")
	}

	if parsed.Fragment != "" {
		return nil, errors.New("jmap: resolver URL must not include a fragment")
	}

	return &StaticResolver{
		url: &url.URL{
			Scheme: parsed.Scheme,
			Host:   parsed.Host,
			Path:   wellKnownPath,
		},
	}, nil
}

// StaticResolver is a Resolver that returns a fixed JMAP session URL.
// It performs no network lookups; the URL is fully determined at construction time.
type StaticResolver struct {
	url *url.URL
}

// Resolve returns the pre-configured JMAP session URL.
// The context is accepted for interface compatibility but is not used.
func (sr *StaticResolver) Resolve(_ context.Context) (*url.URL, error) {
	return sr.url, nil
}
