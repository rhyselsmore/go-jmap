package mail

import (
	"encoding/json"
	"time"

	"github.com/rhyselsmore/go-jmap"
	"github.com/rhyselsmore/go-jmap/spec/patch"
)

// EmailQuery represents a JMAP "Email/query" call (RFC 8621 §4.4).
// It returns a list of Email IDs matching the given filter and sort criteria.
type EmailQuery struct {
	CallID       string       `json:"-"`         // client call id, not on the wire
	AccountID    string       `json:"accountId"` // required
	Filter       *EmailFilter `json:"filter,omitempty"`
	Sort         []SortOption `json:"sort,omitempty"`
	Position     int          `json:"position,omitempty"`
	Anchor       string       `json:"anchor,omitempty"`
	AnchorOffset int          `json:"anchorOffset"`
	Limit        int          `json:"limit,omitempty"`

	response EmailQueryResponse
}

func (e *EmailQuery) Name() string { return "Email/query" }
func (e *EmailQuery) ID() string   { return e.CallID }
func (e *EmailQuery) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &e.response)
}

// Response returns the decoded Email/query result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (e *EmailQuery) Response() EmailQueryResponse {
	return e.response
}

// EmailFilter models the "filter" for Email/query.
// This is a practical subset; JMAP allows lots of fields.
type EmailFilter struct {
	// simple matches
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Cc         string `json:"cc,omitempty"`
	Bcc        string `json:"bcc,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Body       string `json:"body,omitempty"`
	Text       string `json:"text,omitempty"`      // full-text-ish
	MailboxID  string `json:"inMailbox,omitempty"` // emails in a single mailbox
	NotMailbox string `json:"notInMailbox,omitempty"`
	// date stuff (ISO 8601 strings usually)
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
	// flags
	HasKeyword    string `json:"hasKeyword,omitempty"`
	NotKeyword    string `json:"notKeyword,omitempty"`
	HasAttachment *bool  `json:"hasAttachment,omitempty"`

	// compound: "AND", "OR", "NOT"
	Operator   string         `json:"operator,omitempty"`   // "AND" | "OR" | "NOT"
	Conditions []*EmailFilter `json:"conditions,omitempty"` // nested filters
}

// EmailQueryResponse is the arguments object returned by "Email/query".
type EmailQueryResponse struct {
	AccountID           string   `json:"accountId"`
	QueryState          string   `json:"queryState"`
	CanCalculateChanges bool     `json:"canCalculateChanges"`
	Position            int      `json:"position"`
	IDs                 []string `json:"ids"`
	Total               *int     `json:"total,omitempty"`
	Limit               *int     `json:"limit,omitempty"`
}

// EmailGet represents a JMAP "Email/get" call (RFC 8621 §4.5).
// Use IDRef to pass a result reference from a preceding Email/query call,
// allowing both to be batched in a single round trip.
type EmailGet struct {
	CallID              string                `json:"-"`             // client call id
	AccountID           string                `json:"accountId"`     // required
	IDs                 []string              `json:"ids,omitempty"` // if empty, server may return all
	IDRef               *jmap.ResultReference `json:"#ids,omitempty"`
	Properties          []string              `json:"properties,omitempty"` // optional projection
	FetchTextBodyValues bool                  `json:"fetchTextBodyValues"`
	FetchHTMLBodyValues bool                  `json:"fetchHTMLBodyValues"`
	FetchAllBodyValues  bool                  `json:"fetchAllBodyValues"`
	response            EmailGetResponse
}

func (e *EmailGet) Name() string { return "Email/get" }
func (e *EmailGet) ID() string   { return e.CallID }
func (e *EmailGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &e.response)
}

// Response returns the decoded Email/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (e *EmailGet) Response() EmailGetResponse {
	return e.response
}

// EmailGetResponse is the arguments object returned by "Email/get".
type EmailGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Email  `json:"list"`
	NotFound  []string `json:"notFound,omitempty"`
}

// Email represents a JMAP Email object (common fields).
// You can extend this over time as you need more from your server.
type Email struct {
	ID            string                    `json:"id"`
	ThreadID      string                    `json:"threadId,omitempty"` // needed for your Message
	MailboxIDs    map[string]bool           `json:"mailboxIds,omitempty"`
	Keywords      map[string]bool           `json:"keywords,omitempty"`
	Size          int64                     `json:"size,omitempty"`
	ReceivedAt    time.Time                 `json:"receivedAt,omitempty"`
	SentAt        string                    `json:"sentAt,omitempty"`
	Subject       string                    `json:"subject,omitempty"`
	From          []EmailAddress            `json:"from,omitempty"`
	To            []EmailAddress            `json:"to,omitempty"`
	CC            []EmailAddress            `json:"cc,omitempty"`
	BCC           []EmailAddress            `json:"bcc,omitempty"`
	ReplyTo       []EmailAddress            `json:"replyTo,omitempty"`
	Preview       string                    `json:"preview,omitempty"`
	HasAttachment *bool                     `json:"hasAttachment,omitempty"`
	BodyStructure *EmailBodyPart            `json:"bodyStructure,omitempty"`
	BodyValues    map[string]EmailBodyValue `json:"bodyValues,omitempty"`

	// ---- extra bits to feed your local Message ----

	// JMAP can return this as "messageId": ["<id@host>"]
	MessageID []string `json:"messageId,omitempty"`

	// JMAP can return these as arrays too
	InReplyTo  []string `json:"inReplyTo,omitempty"`
	References []string `json:"references,omitempty"`

	// headers: you can request specific ones like "header:List-Id"
	// and put them in a map for convenience
	Headers map[string]string `json:"headers,omitempty"`

	// if you request "header:List-Id" specifically, you can map it here too
	ListID string `json:"listId,omitempty"`
}

// EmailAddress represents a JMAP e-mail address object.
type EmailAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// EmailBodyPart represents a node in the MIME body structure tree of an Email.
// Leaf nodes have a BlobID; container parts have Children.
type EmailBodyPart struct {
	PartID      string          `json:"partId,omitempty"`
	BlobID      string          `json:"blobId,omitempty"`
	Size        int64           `json:"size,omitempty"`
	Type        string          `json:"type,omitempty"` // mime type
	Name        string          `json:"name,omitempty"` // filename
	Disposition string          `json:"disposition,omitempty"`
	CID         string          `json:"cid,omitempty"`
	Children    []EmailBodyPart `json:"subParts,omitempty"` // sometimes "subParts" or "parts" depending on server
}

// EmailBodyValue represents a JMAP "bodyValues" entry.
// It contains the decoded text or reference for one body part.
type EmailBodyValue struct {
	Value             string `json:"value,omitempty"` // actual text content
	IsEncodingProblem bool   `json:"isEncodingProblem,omitempty"`
	IsTruncated       bool   `json:"isTruncated,omitempty"`
}

// CollectAttachments recursively walks a body part tree and appends any parts
// that look like attachments (have a BlobID and a disposition of "attachment"
// or a non-empty filename) to out.
func CollectAttachments(p *EmailBodyPart, out *[]EmailBodyPart) {
	if p == nil {
		return
	}
	if p.BlobID != "" && (p.Disposition == "attachment" || p.Name != "") {
		*out = append(*out, *p)
	}
	for i := range p.Children {
		CollectAttachments(&p.Children[i], out)
	}
}

// EmailSet represents a JMAP "Email/set" call (RFC 8621 §4.7).
// It supports creating, updating, and destroying Email objects in a single call.
type EmailSet struct {
	CallID    string `json:"-"`         // client call id, not on the wire
	AccountID string `json:"accountId"` // required

	// client-supplied objects to create, keyed by client id
	Create map[string]*EmailCreate `json:"create,omitempty"`
	// server ids to update -> patch object
	Update map[string]*EmailPatch `json:"update,omitempty"`
	// server ids to destroy
	Destroy []string `json:"destroy,omitempty"`

	response EmailSetResponse
}

func (e *EmailSet) Name() string { return "Email/set" }
func (e *EmailSet) ID() string   { return e.CallID }
func (e *EmailSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &e.response)
}

// Response returns the decoded Email/set result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (e *EmailSet) Response() EmailSetResponse {
	return e.response
}

// EmailCreate is a minimal, typed object you can send in "create".
// Add more fields as your server supports them.
type EmailCreate struct {
	// At minimum you usually want to put the message in at least 1 mailbox.
	MailboxIDs map[string]bool `json:"mailboxIds,omitempty"`

	// Keywords: e.g. "$seen", "$flagged"
	Keywords map[string]bool `json:"keywords,omitempty"`

	// If you're constructing full messages, you'd add "sentAt", "receivedAt",
	// "from", "to", "subject", "bodyStructure", "bodyValues", etc.
	Subject  string `json:"subject,omitempty"`
	TextBody string `json:"textBody,omitempty"` // non-standard shortcut, but handy if you wrap it server-side
}

// EmailPatch represents the update object for a single email in an
// Email/set call, serialized via [patch.Marshal].
//
// By default each map field (MailboxIDs, Keywords) replaces the whole
// server-side property. Wrap a field with [patch.Partial] to switch it to
// per-entry JMAP patch paths (e.g. "keywords/$seen": true), so only the
// entries you specify are touched. [patch.Value] fields (Subject) are emitted
// as direct property replacements when non-absent, and omitted otherwise.
type EmailPatch struct {
	MailboxIDs patch.Map[bool]     `json:"mailboxIds,omitempty"`
	Keywords   patch.Map[bool]     `json:"keywords,omitempty"`
	Subject    patch.Value[string] `json:"subject,omitempty"`
}

// MarshalJSON serializes the patch using [patch.Marshal].
func (p *EmailPatch) MarshalJSON() ([]byte, error) {
	return patch.Marshal(p)
}

// EmailSetResponse models the "Email/set" response.
type EmailSetResponse struct {
	AccountID string `json:"accountId"`

	// created objects: client id -> created record
	Created map[string]*EmailSetCreated `json:"created,omitempty"`

	// updated ids -> null or object (we'll just type it as *EmailSetUpdated)
	Updated map[string]*EmailSetUpdated `json:"updated,omitempty"`

	// ids that were destroyed
	Destroyed []string `json:"destroyed,omitempty"`

	// notCreated / notUpdated / notDestroyed could be added too
	// if you want full spec.
}

// EmailSetCreated is what you get back for a single created email.
type EmailSetCreated struct {
	ID string `json:"id"`
	// You can add "blobId", "threadId", etc., if your server returns them.
}

// EmailSetUpdated is usually just an empty object, but keep it typed.
type EmailSetUpdated struct{}
