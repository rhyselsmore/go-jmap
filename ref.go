package jmap

// ResultReference is a back-reference to the result of a previous method call
// within the same JMAP request, as defined in RFC 8620 Section 3.7.
// It allows the output of one invocation to be used as input to another,
// enabling dependent calls to be batched in a single round trip.
type ResultReference struct {
	// ResultOf is the call ID of the previous method invocation whose
	// result is being referenced.
	ResultOf string `json:"resultOf"`

	// Name is the method name of the previous invocation, e.g. "Mailbox/query".
	// The server uses this to verify the reference points to the expected method.
	Name string `json:"name"`

	// Path is a JSON Pointer (RFC 6901) into the result object, identifying
	// the value to extract. For example, "/ids/*" references the ids array
	// from a /query response.
	Path string `json:"path"`
}

// Ref creates a ResultReference from an existing Invocation, deriving the
// call ID and method name automatically. The path is a JSON Pointer into
// the referenced invocation's response.
func Ref(inv Invocation, path string) *ResultReference {
	return &ResultReference{
		ResultOf: inv.ID(),
		Name:     inv.Name(),
		Path:     path,
	}
}
