package jmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// SessionGetter is a function that fetches a fresh Session from the server.
// It is passed to [SessionCache.Get] and called only when the cache is empty
// or expired.
type SessionGetter func(ctx context.Context) (Session, error)

// SessionCache manages caching of the JMAP Session object. Implement this
// interface to provide custom caching strategies (e.g. Redis, per-tenant).
type SessionCache interface {
	// Get returns a cached Session, calling fn to fetch a fresh one if needed.
	Get(ctx context.Context, fn SessionGetter) (Session, error)
	// Invalidate discards the cached Session, forcing the next Get to refresh.
	Invalidate(ctx context.Context) error
}

// DefaultSessionCache is an in-memory [SessionCache] that refreshes the
// Session after a configurable TTL. It is safe for concurrent use.
type DefaultSessionCache struct {
	mu      sync.RWMutex
	session Session
	setAt   time.Time
	ttl     time.Duration
}

// NewDefaultSessionCache returns a [DefaultSessionCache] that caches the
// Session for the given TTL duration.
func NewDefaultSessionCache(ttl time.Duration) *DefaultSessionCache {
	return &DefaultSessionCache{
		ttl: ttl,
	}
}

// Get returns the cached Session if it is still within the TTL, otherwise
// calls fn to fetch a fresh one and stores it. Uses a double-checked lock to
// avoid redundant fetches under concurrent access.
func (c *DefaultSessionCache) Get(ctx context.Context, fn SessionGetter) (Session, error) {
	c.mu.RLock()
	if c.session.isSet && time.Since(c.setAt) <= c.ttl {
		s := c.session
		c.mu.RUnlock()
		return s, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session.isSet && time.Since(c.setAt) <= c.ttl {
		return c.session, nil
	}

	s, err := fn(ctx)
	if err != nil {
		return s, err
	}

	c.session = s
	c.setAt = time.Now()
	return s, nil
}

// Invalidate clears the cached Session so the next call to Get fetches a
// fresh one from the server.
func (c *DefaultSessionCache) Invalidate(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = Session{}
	c.setAt = time.Time{}
	return nil
}

// Session represents the JMAP Session object defined in RFC 8620 §2.
type Session struct {
	isSet bool

	// Capabilities advertises the capabilities the server supports,
	// keyed by capability URI (e.g., "urn:ietf:params:jmap:core").
	Capabilities map[Capability]json.RawMessage `json:"capabilities"`

	// Accounts lists accounts the user has access to, keyed by account ID.
	Accounts map[string]Account `json:"accounts"`

	// PrimaryAccounts maps each capability to the default account ID
	// for that capability.
	PrimaryAccounts map[Capability]string `json:"primaryAccounts"`

	// Username is the user’s primary identifier for this session.
	Username string `json:"username"`

	// APIUrl is the endpoint used for all JMAP method calls (POST).
	APIURL string `json:"apiUrl"`

	// DownloadURL is a template for downloading blobs.
	// Replace {accountId} and {blobId}.
	DownloadURL string `json:"downloadUrl"`

	// UploadURL is a template for uploading blobs.
	// Replace {accountId} and {blobId}.
	UploadURL string `json:"uploadUrl"`

	// EventSourceURL is the long-poll URL for push changes.
	EventSourceURL string `json:"eventSourceUrl"`

	// State is a string used to detect when the session object changes.
	State string `json:"state"`

	// Extensions can contain any unrecognized or server-specific fields.
	Extensions map[string]any `json:"-"`
}

// Account represents a single JMAP account.
type Account struct {
	Name                string                         `json:"name"`
	IsPersonal          bool                           `json:"isPersonal"`
	IsReadOnly          bool                           `json:"isReadOnly"`
	AccountCapabilities map[Capability]json.RawMessage `json:"accountCapabilities"`
}

// GetSession returns the cached JMAP session, fetching it from the server if
// the cache is empty or expired.
func (cl *Client) GetSession(ctx context.Context) (Session, error) {
	return cl.session.Get(ctx, func(ctx context.Context) (Session, error) {
		sessionURL, err := cl.resolveSessionURL(ctx)
		if err != nil {
			return Session{}, err
		}

		return cl.fetchSession(ctx, sessionURL)
	})
}

// fetchSession performs an authenticated GET request to the JMAP session URL
// and decodes the response into a Session.
func (cl *Client) fetchSession(ctx context.Context, u *url.URL) (Session, error) {
	var sess Session

	req, err := cl.newRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return sess, err
	}

	resp, err := cl.http.Do(req)
	if err != nil {
		return sess, fmt.Errorf("jmap: fetch session error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return sess, fmt.Errorf("jmap: session request failed: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return sess, fmt.Errorf("jmap: decode session json: %w", err)
	}

	sess.isSet = true

	return sess, nil
}

// resolveSessionURL resolves and caches the JMAP session URL.
// The resolver is called at most once; subsequent calls return the cached result.
func (cl *Client) resolveSessionURL(ctx context.Context) (*url.URL, error) {
	// Fast path: return cached URL under read lock.
	cl.mu.RLock()
	if cl.sessionURL != nil {
		cl.mu.RUnlock()
		return cl.sessionURL, nil
	}
	cl.mu.RUnlock()

	// Slow path: acquire write lock and resolve.
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Re-check after acquiring write lock, another goroutine may have resolved.
	if cl.sessionURL != nil {
		return cl.sessionURL, nil
	}

	u, err := cl.resolver.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	cl.sessionURL = u
	return u, nil
}
