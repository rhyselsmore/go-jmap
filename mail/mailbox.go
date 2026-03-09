package mail

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap"
	"github.com/rhyselsmore/go-jmap/protocol/patch"
)

// MailboxRole identifies the purpose of a mailbox. It is a string type so
// callers can define their own roles beyond the IANA-registered values below.
//
//	custom := mail.MailboxRole("my-custom-role")
type MailboxRole string

const (
	// Standard roles defined in RFC 8621 §2 and registered in the
	// IANA "IMAP Mailbox Name Attributes" registry.

	RoleAll        MailboxRole = "all"        // RFC 8621 §2 — all mail (virtual, subset of \All)
	RoleArchive    MailboxRole = "archive"    // RFC 8621 §2 — archived messages
	RoleDrafts     MailboxRole = "drafts"     // RFC 8621 §2 — draft compositions
	RoleFlagged    MailboxRole = "flagged"    // RFC 8621 §2 — flagged/starred messages
	RoleImportant  MailboxRole = "important"  // RFC 8621 §2 — server-determined important messages
	RoleInbox      MailboxRole = "inbox"      // RFC 8621 §2 — incoming mail
	RoleJunk       MailboxRole = "junk"       // RFC 8621 §2 — spam/junk
	RoleSent       MailboxRole = "sent"       // RFC 8621 §2 — sent messages
	RoleSubscribed MailboxRole = "subscribed" // RFC 8621 §2 — subscribed mailboxes (virtual)
	RoleTrash      MailboxRole = "trash"      // RFC 8621 §2 — deleted messages
)

// Mailbox represents a JMAP Mailbox object as defined in RFC 8621 §2.
// It holds metadata about a mailbox (folder), including counts, permissions,
// and server-specific extension fields.
type Mailbox struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Role               *MailboxRole    `json:"role,omitempty"`
	ParentID           *string         `json:"parentId,omitempty"`
	IsSubscribed       bool            `json:"isSubscribed"`
	IsCollapsed        bool            `json:"isCollapsed"`
	AutoPurge          bool            `json:"autoPurge"`
	AutoLearn          bool            `json:"autoLearn"`
	LearnAsSpam        bool            `json:"learnAsSpam"`
	SuppressDuplicates bool            `json:"suppressDuplicates"`
	PurgeOlderThanDays *int            `json:"purgeOlderThanDays,omitempty"`
	SortOrder          *int            `json:"sortOrder,omitempty"`
	Hidden             *int            `json:"hidden,omitempty"`
	TotalEmails        *int            `json:"totalEmails,omitempty"`
	UnreadEmails       *int            `json:"unreadEmails,omitempty"`
	TotalThreads       *int            `json:"totalThreads,omitempty"`
	UnreadThreads      *int            `json:"unreadThreads,omitempty"`
	IdentityRef        *string         `json:"identityRef,omitempty"`
	MyRights           map[string]bool `json:"myRights,omitempty"`
	Sort               []SortOption    `json:"sort,omitempty"`
}

// MailboxQueryResponse is the arguments object returned by "Mailbox/query".
type MailboxQueryResponse struct {
	AccountID           string   `json:"accountId"`           // the account the query ran on
	QueryState          string   `json:"queryState"`          // state string representing current results
	CanCalculateChanges bool     `json:"canCalculateChanges"` // if true, you can call QueryChanges
	Position            int      `json:"position"`            // index of the first result (for pagination)
	IDs                 []string `json:"ids,omitempty"`       // list of mailbox ids that match
	// Optional fields depending on request options:
	Total *int `json:"total,omitempty"` // total number of results, if requested
	Limit *int `json:"limit,omitempty"` // limit applied, if present
}

// MailboxQuery represents a JMAP "Mailbox/query" call.
type MailboxQuery struct {
	CallID    string                `json:"-"` // not on the wire
	AccountID string                `json:"accountId"`
	Filter    *MailboxFilter        `json:"filter,omitempty"`
	Sort      []SortOption          `json:"sort,omitempty"`
	Position  int                   `json:"position,omitempty"`
	Limit     int                   `json:"limit,omitempty"`
	response  *MailboxQueryResponse `json:"-"`
}

func (mb *MailboxQuery) ID() string   { return mb.CallID }
func (mb *MailboxQuery) Name() string { return "Mailbox/query" }
func (mb *MailboxQuery) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &mb.response)
}

// Response returns the decoded Mailbox/query result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (mb *MailboxQuery) Response() *MailboxQueryResponse { return mb.response }

// MailboxFilter represents the "filter" argument to Mailbox/query.
// You can combine multiple criteria with AND/OR/NOT.
type MailboxFilter struct {
	// match specific fields
	Name         string      `json:"name,omitempty"`
	ParentID     *string     `json:"parentId,omitempty"`
	Role         MailboxRole `json:"role,omitempty"`
	HasAnyRole   *bool       `json:"hasAnyRole,omitempty"`
	IsSubscribed *bool       `json:"isSubscribed,omitempty"`

	// compound logic
	Operator   string           `json:"operator,omitempty"` // "AND", "OR", "NOT"
	Conditions []*MailboxFilter `json:"conditions,omitempty"`
}

// SortOption describes one item in the "sort" array.
type SortOption struct {
	Property    string `json:"property"`              // e.g. "name", "role"
	IsAscending bool   `json:"isAscending,omitempty"` // defaults true if omitted
	Collation   string `json:"collation,omitempty"`   // e.g. "i;unicode-casemap"
}

// MailboxGet represents a JMAP "Mailbox/get" call (RFC 8621 §2.5).
// Use IDRef to pass a result reference from a preceding Mailbox/query call,
// allowing both to be batched in a single round trip.
type MailboxGet struct {
	CallID     string                `json:"-"`             // client call id, not on the wire
	AccountID  string                `json:"accountId"`     // required
	IDs        []string              `json:"ids,omitempty"` // if empty, server may return all
	IDRef      *jmap.ResultReference `json:"#ids,omitempty"`
	Properties []string              `json:"properties,omitempty"`
	response   *MailboxGetResponse   `json:"-"`
}

func (m *MailboxGet) Name() string { return "Mailbox/get" }
func (m *MailboxGet) ID() string   { return m.CallID }
func (mb *MailboxGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &mb.response)
}

// Response returns the decoded Mailbox/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (mb *MailboxGet) Response() *MailboxGetResponse { return mb.response }

// MailboxGetResponse is the arguments object returned by "Mailbox/get".
type MailboxGetResponse struct {
	AccountID string    `json:"accountId"`
	State     string    `json:"state"`
	List      []Mailbox `json:"list"`
	NotFound  []string  `json:"notFound,omitempty"`
}

// MailboxChanges represents a JMAP "Mailbox/changes" call (RFC 8621 §2.3,
// RFC 8620 §5.2). It returns the IDs of mailboxes that have been created,
// updated, or destroyed since the given state.
type MailboxChanges struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// SinceState is the state string from a previous Mailbox/get response.
	// The server returns changes that occurred after this state.
	SinceState string `json:"sinceState"`

	// MaxChanges limits the number of IDs returned. The server MAY return
	// fewer. If more changes exist, HasMoreChanges will be true.
	MaxChanges *int `json:"maxChanges,omitempty"`

	response *MailboxChangesResponse `json:"-"`
}

func (m *MailboxChanges) Name() string { return "Mailbox/changes" }
func (m *MailboxChanges) ID() string   { return m.CallID }
func (m *MailboxChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &m.response)
}

// Response returns the decoded Mailbox/changes result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (m *MailboxChanges) Response() *MailboxChangesResponse { return m.response }

// MailboxChangesResponse is the arguments object returned by "Mailbox/changes".
type MailboxChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}

// MailboxQueryChanges represents a JMAP "Mailbox/queryChanges" call
// (RFC 8621 §2.5, RFC 8620 §5.6). It returns the changes to a query's result
// set since a given query state, expressed as removed IDs and added items
// (with their new index positions).
//
// The Filter and Sort fields must match the original Mailbox/query call whose
// state you are diffing against.
type MailboxQueryChanges struct {
	CallID    string         `json:"-"`
	AccountID string         `json:"accountId"`
	Filter    *MailboxFilter `json:"filter,omitempty"`
	Sort      []SortOption   `json:"sort,omitempty"`

	// SinceQueryState is the queryState string from a previous Mailbox/query
	// response.
	SinceQueryState string `json:"sinceQueryState"`

	// MaxChanges limits the total number of removed + added entries returned.
	// If more changes exist the server returns a tooManyChanges error.
	MaxChanges *int `json:"maxChanges,omitempty"`

	// UpToID, if set, tells the server to only return changes up to (and
	// including) this ID in the result set. IDs after it are omitted.
	UpToID string `json:"upToId,omitempty"`

	// CalculateTotal requests that the server include the total count of
	// results in the response.
	CalculateTotal bool `json:"calculateTotal,omitempty"`

	response *MailboxQueryChangesResponse `json:"-"`
}

func (m *MailboxQueryChanges) Name() string { return "Mailbox/queryChanges" }
func (m *MailboxQueryChanges) ID() string   { return m.CallID }
func (m *MailboxQueryChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &m.response)
}

// Response returns the decoded Mailbox/queryChanges result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (m *MailboxQueryChanges) Response() *MailboxQueryChangesResponse { return m.response }

// MailboxQueryChangesResponse is the arguments object returned by
// "Mailbox/queryChanges".
type MailboxQueryChangesResponse struct {
	AccountID     string      `json:"accountId"`
	OldQueryState string      `json:"oldQueryState"`
	NewQueryState string      `json:"newQueryState"`
	Removed       []string    `json:"removed"`
	Added         []AddedItem `json:"added"`
	Total         *int        `json:"total,omitempty"`
}

// AddedItem represents a single entry in the "added" array of a
// queryChanges response (RFC 8620 §5.6). It pairs an object ID with its
// new position in the sorted result set.
type AddedItem struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
}

// MailboxSet represents a JMAP "Mailbox/set" call (RFC 8621 §2.6).
// It supports creating, updating, and destroying Mailbox objects in a single
// call. Use IfInState for optimistic concurrency control.
type MailboxSet struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// IfInState, if set, must match the current state string from Mailbox/get.
	// If it does not match, the server rejects the call with a stateMismatch error.
	IfInState string `json:"ifInState,omitempty"`

	Create  map[string]*MailboxCreate `json:"create,omitempty"`
	Update  map[string]*MailboxPatch  `json:"update,omitempty"`
	Destroy []string                  `json:"destroy,omitempty"`

	// OnDestroyRemoveEmails controls what happens to emails in a destroyed
	// mailbox. If true, emails that exist only in the destroyed mailbox are
	// also destroyed. If false (the default), the server rejects the destroy
	// with a mailboxHasEmail error when emails would be orphaned.
	OnDestroyRemoveEmails bool `json:"onDestroyRemoveEmails,omitempty"`

	response *MailboxSetResponse
}

func (m *MailboxSet) Name() string { return "Mailbox/set" }
func (m *MailboxSet) ID() string   { return m.CallID }
func (m *MailboxSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &m.response)
}

// Response returns the decoded Mailbox/set result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (m *MailboxSet) Response() *MailboxSetResponse { return m.response }

// MailboxCreate describes a new mailbox to create in a Mailbox/set call.
// Server-only properties (ID, counts, myRights) must be omitted.
type MailboxCreate struct {
	Name         string       `json:"name"`
	ParentID     *string      `json:"parentId,omitempty"`
	Role         *MailboxRole `json:"role,omitempty"`
	IsSubscribed bool         `json:"isSubscribed,omitempty"`
	SortOrder    int          `json:"sortOrder,omitempty"`
}

// MailboxPatch represents the update object for a single mailbox in a
// Mailbox/set call, serialized via [patch.Marshal].
//
// Each field uses [patch.Value] for three-state semantics: absent (omitted
// from the patch), set (a concrete value), or null (explicit deletion).
type MailboxPatch struct {
	Name         patch.Value[string]      `json:"name"`
	ParentID     patch.Value[string]      `json:"parentId"`
	Role         patch.Value[MailboxRole] `json:"role"`
	IsSubscribed patch.Value[bool]        `json:"isSubscribed"`
	SortOrder    patch.Value[int]         `json:"sortOrder"`
}

// MarshalJSON serializes the patch using [patch.Marshal].
func (p *MailboxPatch) MarshalJSON() ([]byte, error) {
	return patch.Marshal(p)
}

// MailboxSetResponse is the arguments object returned by "Mailbox/set",
// as defined in RFC 8620 §5.3.
type MailboxSetResponse struct {
	AccountID    string                        `json:"accountId"`
	OldState     string                        `json:"oldState"`
	NewState     string                        `json:"newState"`
	Created      map[string]*MailboxSetCreated `json:"created,omitempty"`
	Updated      map[string]*MailboxSetUpdated `json:"updated,omitempty"`
	Destroyed    []string                      `json:"destroyed,omitempty"`
	NotCreated   map[string]*SetError          `json:"notCreated,omitempty"`
	NotUpdated   map[string]*SetError          `json:"notUpdated,omitempty"`
	NotDestroyed map[string]*SetError          `json:"notDestroyed,omitempty"`
}

// MailboxSetCreated contains the server-assigned properties returned for a
// newly created mailbox.
type MailboxSetCreated struct {
	ID string `json:"id"`
}

// MailboxSetUpdated may contain any server-side changes beyond what was
// requested. Typically empty.
type MailboxSetUpdated struct{}

// SetError represents an error for a single create, update, or destroy
// operation within a /set response, as defined in RFC 8620 §5.3.
type SetError struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
