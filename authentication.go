package jmap

import (
	"errors"
	"net/http"
	"strings"
)

// Authenticator signs outgoing HTTP requests, e.g. by adding an Authorization
// header. Implement this interface to support custom authentication schemes.
type Authenticator interface {
	// Authenticate mutates req to add authentication credentials.
	Authenticate(req *http.Request) error
}

// WithBearerTokenAuthentication is a convenience [ClientOpt] that creates a
// [BearerTokenAuthenticator] from the given token and sets it on the client.
func WithBearerTokenAuthentication(token string) ClientOpt {
	return func(c *Client) error {
		var err error
		c.authn, err = NewBearerTokenAuthenticator(token)
		return err
	}
}

// NewBearerTokenAuthenticator returns a [BearerTokenAuthenticator] for the
// given token. Returns an error if the token is empty or whitespace-only.
func NewBearerTokenAuthenticator(token string) (*BearerTokenAuthenticator, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("jmap: bearer token must not be empty")
	}
	return &BearerTokenAuthenticator{
		token: token,
	}, nil
}

// BearerTokenAuthenticator implements [Authenticator] using a static Bearer
// token. Use [NewBearerTokenAuthenticator] or [WithBearerTokenAuthentication]
// to construct one.
type BearerTokenAuthenticator struct {
	token string
}

func (b *BearerTokenAuthenticator) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return nil
}
