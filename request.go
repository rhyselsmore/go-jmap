package jmap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Invocation represents a single JMAP method invocation in a request.
type Invocation interface {
	// Name returns the JMAP method name, e.g. "Mailbox/query".
	Name() string

	// ID returns the client-specified call id (the 3rd element in the array).
	// If it returns "", the request will generate one automatically.
	ID() string

	// DecodeResponse decodes the raw JSON response into the invocation's
	// response type. The provided json.RawMessage is safe to mutate.
	DecodeResponse(b json.RawMessage) error
}

// Request is the top-level JMAP request envelope sent to the server.
type Request struct {
	Using       []Capability
	MethodCalls []Invocation

	// ids maps the resolved call ID to its invocation, built during Add.
	ids []resolvedCall
}

// resolvedCall pairs an invocation with its final, deduplicated call ID.
type resolvedCall struct {
	id  string
	inv Invocation
}

// NewRequest creates a new request with the given capabilities.
func NewRequest(using ...Capability) *Request {
	return &Request{
		Using:       using,
		MethodCalls: make([]Invocation, 0),
	}
}

// Add appends a method call to the request and assigns a unique call ID.
func (r *Request) Add(inv Invocation) {
	id := inv.ID()
	if id == "" {
		id = fmt.Sprintf("c%d", len(r.MethodCalls)+1)
	}

	// Ensure uniqueness
	seen := make(map[string]struct{}, len(r.ids))
	for _, rc := range r.ids {
		seen[rc.id] = struct{}{}
	}
	orig := id
	suffix := 0
	for {
		if _, exists := seen[id]; !exists {
			break
		}
		suffix++
		id = fmt.Sprintf("%s.%d", orig, suffix)
	}

	r.ids = append(r.ids, resolvedCall{id: id, inv: inv})
	r.MethodCalls = append(r.MethodCalls, inv)
}

// MarshalJSON converts the request into the JMAP wire format:
//
//	{
//	  "using": [...],
//	  "methodCalls": [
//	    ["Mailbox/query", {...}, "c1"],
//	    ...
//	  ]
//	}
func (r *Request) MarshalJSON() ([]byte, error) {
	type encoded struct {
		Using       []Capability `json:"using"`
		MethodCalls [][]any      `json:"methodCalls"`
	}

	enc := encoded{
		Using:       r.Using,
		MethodCalls: make([][]any, 0, len(r.ids)),
	}

	for _, rc := range r.ids {
		enc.MethodCalls = append(enc.MethodCalls, []any{
			rc.inv.Name(),
			rc.inv,
			rc.id,
		})
	}

	return json.Marshal(enc)
}

// lookup returns the invocation associated with the given call ID.
func (r *Request) lookup(id string) (Invocation, bool) {
	for _, rc := range r.ids {
		if rc.id == id {
			return rc.inv, true
		}
	}
	return nil, false
}

// Do executes a JMAP request against the server's API URL, decodes the
// response, and correlates each method response back to its originating
// invocation via [Response.correlate]. Returns an error for non-2xx HTTP
// status codes, JSON decode failures, or correlation errors.
func (cl *Client) Do(ctx context.Context, req *Request) (Response, error) {
	var resp Response

	sess, err := cl.GetSession(ctx)
	if err != nil {
		return resp, err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return resp, fmt.Errorf("jmap: marshal request: %w", err)
	}

	httpReq, err := cl.newRequest(ctx, http.MethodPost, sess.APIURL, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := cl.http.Do(httpReq)
	if err != nil {
		return resp, fmt.Errorf("jmap: request error: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		return resp, fmt.Errorf("jmap: request failed: %s", httpResp.Status)
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return resp, fmt.Errorf("jmap: decode response json: %w", err)
	}

	if err := resp.correlate(req); err != nil {
		return resp, err
	}

	return resp, nil
}
