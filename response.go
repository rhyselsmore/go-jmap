package jmap

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Response is the top-level JMAP response envelope.
type Response struct {
	MethodResponses []MethodResponse `json:"methodResponses"`
	SessionState    string           `json:"sessionState"`
}

// MethodResponse represents a single JMAP method response triple:
// [ "Method/name", {args}, "clientCallId" ]
type MethodResponse struct {
	Name string          // e.g. "Mailbox/get"
	Args json.RawMessage // raw JSON object of the method's result
	ID   string          // clientCallId, e.g. "c1"
}

// InvocationError represents a server-side error returned for a single
// JMAP method invocation.
type InvocationError struct {
	CallID string // the call ID of the failed invocation
	Type   string `json:"type"`
	Detail string `json:"description,omitempty"`
}

func (e *InvocationError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("jmap: invocation %q error: %s: %s", e.CallID, e.Type, e.Detail)
	}
	return fmt.Sprintf("jmap: invocation %q error: %s", e.CallID, e.Type)
}

// UnmarshalJSON implements custom decoding for the JMAP response envelope.
func (r *Response) UnmarshalJSON(b []byte) error {
	var raw struct {
		MethodResponses [][]json.RawMessage `json:"methodResponses"`
		SessionState    string              `json:"sessionState"`
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("jmap: decode response envelope: %w", err)
	}

	r.SessionState = raw.SessionState
	r.MethodResponses = make([]MethodResponse, 0, len(raw.MethodResponses))

	for i, triple := range raw.MethodResponses {
		if len(triple) != 3 {
			return fmt.Errorf("jmap: method response %d: expected 3 elements, got %d", i, len(triple))
		}

		var name, id string
		if err := json.Unmarshal(triple[0], &name); err != nil {
			return fmt.Errorf("jmap: method response %d: decode name: %w", i, err)
		}
		if err := json.Unmarshal(triple[2], &id); err != nil {
			return fmt.Errorf("jmap: method response %d: decode call id: %w", i, err)
		}

		r.MethodResponses = append(r.MethodResponses, MethodResponse{
			Name: name,
			Args: triple[1],
			ID:   id,
		})
	}

	return nil
}

// correlate matches each method response to its original invocation and
// decodes the result. If any invocation returned a server-side error,
// it is collected and returned as a joined error.
func (r *Response) correlate(req *Request) error {
	var errs []error

	for _, mr := range r.MethodResponses {
		inv, ok := req.lookup(mr.ID)
		if !ok {
			continue
		}

		// Handle JMAP-level invocation errors.
		if mr.Name == "error" {
			invErr := &InvocationError{CallID: mr.ID}
			if err := json.Unmarshal(mr.Args, invErr); err != nil {
				errs = append(errs, fmt.Errorf("jmap: decode error for invocation %q: %w", mr.ID, err))
				continue
			}
			errs = append(errs, invErr)
			continue
		}

		if err := inv.DecodeResponse(mr.Args); err != nil {
			errs = append(errs, fmt.Errorf("jmap: decode response for invocation %q (%s): %w", mr.ID, mr.Name, err))
		}
	}

	return errors.Join(errs...)
}
