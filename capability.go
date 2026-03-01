package jmap

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Capability is a JMAP capability identifier, typically a URN such as
// "urn:ietf:params:jmap:core" or "urn:ietf:params:jmap:mail". Capabilities
// are used as keys in the Session and Account objects to advertise server
// support and constraints, and are passed in the "using" field of requests.
type Capability string

// GetCapabilities decodes the server-level capability object for the given
// capability identifier from the Session. The type parameter T should match
// the expected shape of the capability object as defined by the relevant RFC.
// Returns an error if the capability is not present or cannot be decoded.
func GetCapabilities[T any](s Session, c Capability) (T, error) {
	var cap T
	raw, ok := s.Capabilities[c]
	if !ok {
		return cap, errors.New("jmap: capability not present")
	}
	if err := json.Unmarshal(raw, &cap); err != nil {
		return cap, fmt.Errorf("jmap: decode capability: %w", err)
	}
	return cap, nil
}

// GetAccountCapabilities decodes the account-level capability object for the
// given capability identifier from an Account. Account-level capabilities
// describe per-account constraints and permissions, which may differ from the
// server-level capabilities. Returns an error if the capability is not present
// or cannot be decoded.
func GetAccountCapabilities[T any](acc Account, c Capability) (T, error) {
	var cap T
	raw, ok := acc.AccountCapabilities[c]
	if !ok {
		return cap, errors.New("jmap: capability not present")
	}
	if err := json.Unmarshal(raw, &cap); err != nil {
		return cap, fmt.Errorf("jmap: decode account capability: %w", err)
	}
	return cap, nil
}
