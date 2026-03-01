package jmap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// roundTripperFunc adapts a plain function into an [http.RoundTripper],
// used to inject authentication into every outgoing HTTP request.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

// ClientOpt is a functional option applied to a [Client] during construction.
type ClientOpt func(*Client) error

// WithHTTPClient configures the [Client] to use the provided *http.Client
// instead of the default. The client is shallow-cloned to avoid mutating
// the caller's value.
func WithHTTPClient(cl *http.Client) ClientOpt {
	return func(c *Client) error {
		clone := *cl
		c.http = &clone
		return nil
	}
}

// WithAuthenticator sets the [Authenticator] used to sign every outgoing
// request. This option is required; [NewClient] will return an error if it
// is not provided.
func WithAuthenticator(authn Authenticator) ClientOpt {
	return func(c *Client) error {
		c.authn = authn
		return nil
	}
}

// NewClient constructs a [Client] by applying the given options and then
// validating that all required fields (resolver, authenticator) are set.
func NewClient(opts ...ClientOpt) (*Client, error) {
	cl := &Client{}

	// Apply Opts
	if err := cl.applyOpts(opts...); err != nil {
		return nil, err
	}

	if err := cl.configure(); err != nil {
		return nil, err
	}

	return cl, nil
}

// Client is a JMAP client. Use [NewClient] to construct one.
type Client struct {
	http     *http.Client
	resolver Resolver
	session  SessionCache
	authn    Authenticator

	mu         sync.RWMutex
	sessionURL *url.URL
}

func (cl *Client) applyOpts(opts ...ClientOpt) error {
	for _, opt := range opts {
		if err := opt(cl); err != nil {
			return err
		}
	}
	return nil
}

func (cl *Client) configure() error {
	// Resolver is required
	if cl.resolver == nil {
		return errors.New("jmap: resolver must be provided using WithResolver")
	}

	// Use Default Session Cache - 5 minute TTL
	if cl.session == nil {
		cl.session = NewDefaultSessionCache(time.Second * 300)
	}

	// Authenticator is required
	if cl.authn == nil {
		return errors.New("jmap: authenticator must be provided using WithAuthenticator")
	}

	// Set Default Client
	if cl.http == nil {
		cl.http = &http.Client{
			Timeout: http.DefaultClient.Timeout,
		}
	}

	// Important: Capture the existing transport or use the default
	innerTransport := cl.http.Transport
	if innerTransport == nil {
		innerTransport = http.DefaultTransport
	}

	cl.http.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())
			if err := cl.authn.Authenticate(req); err != nil {
				return nil, err
			}
			return innerTransport.RoundTrip(req)
		},
	)

	return nil
}

// newRequest creates an *http.Request with the given method, URL, and body.
// If ctx is nil it falls back to context.Background().
func (cl *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("jmap: new request: %w", err)
	}

	return req, nil
}
