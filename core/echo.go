package core

import "encoding/json"

// Echo represents a JMAP "Core/echo" call (RFC 8620 §4). The server returns
// exactly the same arguments it receives, making it useful for verifying
// authenticated connectivity to the API endpoint.
//
// The type parameter T defines the shape of the arguments and response,
// avoiding untyped map[string]any round-trips.
type Echo[T any] struct {
	CallID string `json:"-"`
	Args   T      `json:"-"`

	response T `json:"-"`
}

func (e *Echo[T]) Name() string { return "Core/echo" }
func (e *Echo[T]) ID() string   { return e.CallID }

// MarshalJSON serializes the arguments as the invocation body.
func (e *Echo[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Args)
}

func (e *Echo[T]) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &e.response)
}

// Response returns the echoed arguments. It is only populated after the
// request has been executed via [jmap.Client.Do].
func (e *Echo[T]) Response() T { return e.response }
