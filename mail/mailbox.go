package mail

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap"
)

type Mailbox struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Role               *string         `json:"role,omitempty"`
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
func (mb *MailboxQuery) Response() *MailboxQueryResponse { return mb.response }

// MailboxFilter represents the "filter" argument to Mailbox/query.
// You can combine multiple criteria with AND/OR/NOT.
type MailboxFilter struct {
	// match specific fields
	Name         string  `json:"name,omitempty"`
	ParentID     *string `json:"parentId,omitempty"`
	Role         string  `json:"role,omitempty"`
	HasAnyRole   *bool   `json:"hasAnyRole,omitempty"`
	IsSubscribed *bool   `json:"isSubscribed,omitempty"`

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

// MailboxGet represents a JMAP "Mailbox/get" call.
type MailboxGet struct {
	CallID    string                `json:"-"`             // client call id, not on the wire
	AccountID string                `json:"accountId"`     // required
	IDs       []string              `json:"ids,omitempty"` // if empty, server may return all
	IDRef     *jmap.ResultReference `json:"#ids,omitempty"`
	// Properties lets you limit the fields returned (recommended)
	Properties []string            `json:"properties,omitempty"`
	response   *MailboxGetResponse `json:"-"`
}

func (m *MailboxGet) Name() string { return "Mailbox/get" }
func (m *MailboxGet) ID() string   { return m.CallID }
func (mb *MailboxGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &mb.response)
}
func (mb *MailboxGet) Response() *MailboxGetResponse { return mb.response }

// MailboxGetResponse is the arguments object returned by "Mailbox/get".
type MailboxGetResponse struct {
	AccountID string    `json:"accountId"`
	State     string    `json:"state"`
	List      []Mailbox `json:"list"`
	NotFound  []string  `json:"notFound,omitempty"`
}
