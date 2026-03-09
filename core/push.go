package core

import "encoding/json"

// PushSubscription represents a JMAP PushSubscription object as defined in
// RFC 8620 §7.2. It registers a URL where the server will POST notifications
// when data changes. PushSubscriptions are per-user, not per-account.
type PushSubscription struct {
	ID               string            `json:"id"`
	DeviceClientID   string            `json:"deviceClientId"`
	URL              string            `json:"url"`
	Keys             *PushKeys         `json:"keys,omitempty"`
	VerificationCode string            `json:"verificationCode,omitempty"`
	Expires          *string           `json:"expires,omitempty"`
	Types            []string          `json:"types,omitempty"`
}

// PushKeys contains the encryption keys for a PushSubscription, enabling
// end-to-end encrypted push notifications per RFC 8291 (Web Push Encryption).
type PushKeys struct {
	// P256DH is the ECDH public key (base64url-encoded) from the client.
	P256DH string `json:"p256dh"`
	// Auth is the authentication secret (base64url-encoded) from the client.
	Auth string `json:"auth"`
}

// PushSubscriptionGet represents a JMAP "PushSubscription/get" call
// (RFC 8620 §7.2.1). Unlike most /get methods, this takes no accountId
// because PushSubscriptions are user-level, not account-level.
type PushSubscriptionGet struct {
	CallID     string   `json:"-"`
	IDs        []string `json:"ids,omitempty"`
	Properties []string `json:"properties,omitempty"`

	response *PushSubscriptionGetResponse `json:"-"`
}

func (p *PushSubscriptionGet) Name() string { return "PushSubscription/get" }
func (p *PushSubscriptionGet) ID() string   { return p.CallID }
func (p *PushSubscriptionGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &p.response)
}

// Response returns the decoded PushSubscription/get result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (p *PushSubscriptionGet) Response() *PushSubscriptionGetResponse { return p.response }

// PushSubscriptionGetResponse is the arguments object returned by
// "PushSubscription/get".
type PushSubscriptionGetResponse struct {
	List     []PushSubscription `json:"list"`
	NotFound []string           `json:"notFound,omitempty"`
}

// PushSubscriptionSet represents a JMAP "PushSubscription/set" call
// (RFC 8620 §7.2.2). Like /get, it takes no accountId.
type PushSubscriptionSet struct {
	CallID  string                                `json:"-"`
	Create  map[string]*PushSubscriptionCreate    `json:"create,omitempty"`
	Update  map[string]*PushSubscriptionUpdate    `json:"update,omitempty"`
	Destroy []string                              `json:"destroy,omitempty"`

	response *PushSubscriptionSetResponse `json:"-"`
}

func (p *PushSubscriptionSet) Name() string { return "PushSubscription/set" }
func (p *PushSubscriptionSet) ID() string   { return p.CallID }
func (p *PushSubscriptionSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &p.response)
}

// Response returns the decoded PushSubscription/set result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (p *PushSubscriptionSet) Response() *PushSubscriptionSetResponse { return p.response }

// PushSubscriptionCreate describes a new PushSubscription to register.
type PushSubscriptionCreate struct {
	DeviceClientID   string    `json:"deviceClientId"`
	URL              string    `json:"url"`
	Keys             *PushKeys `json:"keys,omitempty"`
	VerificationCode string    `json:"verificationCode,omitempty"`
	Expires          *string   `json:"expires,omitempty"`
	Types            []string  `json:"types,omitempty"`
}

// PushSubscriptionUpdate describes updates to an existing PushSubscription.
// Only VerificationCode, Expires, and Types are mutable.
type PushSubscriptionUpdate struct {
	VerificationCode *string  `json:"verificationCode,omitempty"`
	Expires          *string  `json:"expires,omitempty"`
	Types            []string `json:"types,omitempty"`
}

// PushSubscriptionSetResponse is the arguments object returned by
// "PushSubscription/set".
type PushSubscriptionSetResponse struct {
	Created      map[string]*PushSubscriptionSetCreated `json:"created,omitempty"`
	Updated      map[string]any                         `json:"updated,omitempty"`
	Destroyed    []string                               `json:"destroyed,omitempty"`
	NotCreated   map[string]*PushSubscriptionSetError   `json:"notCreated,omitempty"`
	NotUpdated   map[string]*PushSubscriptionSetError   `json:"notUpdated,omitempty"`
	NotDestroyed map[string]*PushSubscriptionSetError   `json:"notDestroyed,omitempty"`
}

// PushSubscriptionSetCreated contains the server-assigned properties returned
// for a newly created PushSubscription.
type PushSubscriptionSetCreated struct {
	ID      string  `json:"id"`
	Expires *string `json:"expires,omitempty"`
}

// PushSubscriptionSetError represents an error for a single create, update, or
// destroy operation within a PushSubscription/set response.
type PushSubscriptionSetError struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
