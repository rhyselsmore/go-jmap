package core

import "github.com/rhyselsmore/go-jmap"

// Capability is the JMAP capability identifier for the core protocol, as
// defined in RFC 8620. This capability MUST be present in every JMAP Session
// and describes the server's fundamental limits for requests, uploads, and
// object operations.
const Capability jmap.Capability = "urn:ietf:params:jmap:core"

// Capabilities contains the server-level constraints for the core JMAP
// protocol, as defined in RFC 8620 Section 2.
type Capabilities struct {
	// MaxSizeUpload is the maximum file size, in octets, that the server
	// will accept for a single file upload. Suggested minimum: 50,000,000.
	MaxSizeUpload int `json:"maxSizeUpload"`

	// MaxConcurrentUpload is the maximum number of concurrent requests the
	// server will accept to the upload endpoint. Suggested minimum: 4.
	MaxConcurrentUpload int `json:"maxConcurrentUpload"`

	// MaxSizeRequest is the maximum size, in octets, that the server will
	// accept for a single request to the API endpoint. Suggested minimum: 10,000,000.
	MaxSizeRequest int `json:"maxSizeRequest"`

	// MaxConcurrentRequests is the maximum number of concurrent requests the
	// server will accept to the API endpoint. Suggested minimum: 4.
	MaxConcurrentRequests int `json:"maxConcurrentRequests"`

	// MaxCallsInRequest is the maximum number of method calls the server will
	// accept in a single request to the API endpoint. Suggested minimum: 16.
	MaxCallsInRequest int `json:"maxCallsInRequest"`

	// MaxObjectsInGet is the maximum number of objects that the client may
	// request in a single /get type method call. Suggested minimum: 500.
	MaxObjectsInGet int `json:"maxObjectsInGet"`

	// MaxObjectsInSet is the maximum number of objects the client may send to
	// create, update, or destroy in a single /set type method call. This is the
	// combined total across all three operations. Suggested minimum: 500.
	MaxObjectsInSet int `json:"maxObjectsInSet"`
}

// GetCapabilities decodes the server-level core capability from the Session.
// The core capability is always present in a valid JMAP Session and contains
// the server's fundamental limits that clients should respect.
func GetCapabilities(s jmap.Session) (*Capabilities, error) {
	return jmap.GetCapabilities[*Capabilities](s, Capability)
}
