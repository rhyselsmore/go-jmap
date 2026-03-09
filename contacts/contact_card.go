package contacts

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap"
	"github.com/rhyselsmore/go-jmap/protocol/patch"
)

// ContactCard represents a JMAP ContactCard object as defined in RFC 9610 §4.
// It wraps a JSContact Card (RFC 9553) with JMAP-specific properties.
//
// Only commonly used JSContact properties are included here; extend this
// struct as needed for your use case.
type ContactCard struct {
	// JMAP-specific properties
	ID             string          `json:"id"`
	AddressBookIDs map[string]bool `json:"addressBookIds,omitempty"`

	// JSContact metadata (RFC 9553 §2.1)
	UID     string `json:"uid,omitempty"`
	Kind    string `json:"kind,omitempty"` // "individual", "group", "org", etc.
	Created string `json:"created,omitempty"`
	Updated string `json:"updated,omitempty"`

	// Names (RFC 9553 §2.2)
	Name      *Name                `json:"name,omitempty"`
	Nicknames map[string]*Nickname `json:"nicknames,omitempty"`

	// Contact info (RFC 9553 §2.3)
	Emails         map[string]*EmailAddress  `json:"emails,omitempty"`
	Phones         map[string]*Phone         `json:"phones,omitempty"`
	OnlineServices map[string]*OnlineService `json:"onlineServices,omitempty"`

	// Addresses (RFC 9553 §2.5)
	Addresses map[string]*Address `json:"addresses,omitempty"`

	// Organizations and titles (RFC 9553 §2.2)
	Organizations map[string]*Organization `json:"organizations,omitempty"`
	Titles        map[string]*Title        `json:"titles,omitempty"`

	// Other
	Notes    map[string]*Note `json:"notes,omitempty"`
	Keywords map[string]bool  `json:"keywords,omitempty"`
}

// Name represents a structured name (RFC 9553 §2.2.1).
type Name struct {
	Components []NameComponent `json:"components,omitempty"`
	Full       string          `json:"full,omitempty"`
	IsOrdered  bool            `json:"isOrdered,omitempty"`
}

// NameComponent represents a single component of a structured name.
type NameComponent struct {
	Kind  string `json:"kind"` // "given", "surname", "surname2", "title", "credential", etc.
	Value string `json:"value"`
}

// Nickname represents a nickname (RFC 9553 §2.2.2).
type Nickname struct {
	Name string `json:"name"`
}

// EmailAddress represents an email address in a ContactCard (RFC 9553 §2.3.1).
type EmailAddress struct {
	Address  string          `json:"address"`
	Contexts map[string]bool `json:"contexts,omitempty"` // "work", "private"
	Pref     int             `json:"pref,omitempty"`     // 1-100, lower is higher priority
	Label    string          `json:"label,omitempty"`
}

// Phone represents a phone number (RFC 9553 §2.3.3).
type Phone struct {
	Number   string          `json:"number"`
	Features map[string]bool `json:"features,omitempty"` // "voice", "fax", "cell", "text", etc.
	Contexts map[string]bool `json:"contexts,omitempty"` // "work", "private"
	Pref     int             `json:"pref,omitempty"`
	Label    string          `json:"label,omitempty"`
}

// OnlineService represents an online service handle (RFC 9553 §2.3.2).
type OnlineService struct {
	Service  string          `json:"service,omitempty"` // e.g. "twitter", "github"
	User     string          `json:"user,omitempty"`
	URI      string          `json:"uri,omitempty"`
	Contexts map[string]bool `json:"contexts,omitempty"`
	Pref     int             `json:"pref,omitempty"`
	Label    string          `json:"label,omitempty"`
}

// Address represents a postal address (RFC 9553 §2.5.1).
type Address struct {
	Components  []AddressComponent `json:"components,omitempty"`
	Full        string             `json:"full,omitempty"`
	IsOrdered   bool               `json:"isOrdered,omitempty"`
	CountryCode string             `json:"countryCode,omitempty"`
	Coordinates string             `json:"coordinates,omitempty"` // "geo:" URI
	TimeZone    string             `json:"timeZone,omitempty"`
	Contexts    map[string]bool    `json:"contexts,omitempty"`
	Pref        int                `json:"pref,omitempty"`
	Label       string             `json:"label,omitempty"`
}

// AddressComponent represents a single component of a structured address.
type AddressComponent struct {
	Kind  string `json:"kind"` // "locality", "region", "country", "postcode", etc.
	Value string `json:"value"`
}

// Organization represents an organization (RFC 9553 §2.2.3).
type Organization struct {
	Name     string          `json:"name,omitempty"`
	Units    []OrgUnit       `json:"units,omitempty"`
	Contexts map[string]bool `json:"contexts,omitempty"`
}

// OrgUnit represents an organizational unit.
type OrgUnit struct {
	Name string `json:"name"`
}

// Title represents a job title or role (RFC 9553 §2.2.4).
type Title struct {
	Name     string          `json:"name"`
	Kind     string          `json:"kind,omitempty"` // "title" or "role"
	Contexts map[string]bool `json:"contexts,omitempty"`
}

// Note represents a free-text note (RFC 9553 §2.8.3).
type Note struct {
	Note string `json:"note"`
}

// ContactCardGet represents a JMAP "ContactCard/get" call (RFC 9610 §4.2).
type ContactCardGet struct {
	CallID     string                  `json:"-"`
	AccountID  string                  `json:"accountId"`
	IDs        []string                `json:"ids,omitempty"`
	IDRef      *jmap.ResultReference   `json:"#ids,omitempty"`
	Properties []string                `json:"properties,omitempty"`
	response   *ContactCardGetResponse `json:"-"`
}

func (c *ContactCardGet) Name() string { return "ContactCard/get" }
func (c *ContactCardGet) ID() string   { return c.CallID }
func (c *ContactCardGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (c *ContactCardGet) Response() *ContactCardGetResponse { return c.response }

// ContactCardGetResponse is the arguments object returned by "ContactCard/get".
type ContactCardGetResponse struct {
	AccountID string        `json:"accountId"`
	State     string        `json:"state"`
	List      []ContactCard `json:"list"`
	NotFound  []string      `json:"notFound,omitempty"`
}

// ContactCardChanges represents a JMAP "ContactCard/changes" call
// (RFC 9610 §4.3, RFC 8620 §5.2).
type ContactCardChanges struct {
	CallID     string                      `json:"-"`
	AccountID  string                      `json:"accountId"`
	SinceState string                      `json:"sinceState"`
	MaxChanges *int                        `json:"maxChanges,omitempty"`
	response   *ContactCardChangesResponse `json:"-"`
}

func (c *ContactCardChanges) Name() string { return "ContactCard/changes" }
func (c *ContactCardChanges) ID() string   { return c.CallID }
func (c *ContactCardChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/changes result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (c *ContactCardChanges) Response() *ContactCardChangesResponse { return c.response }

// ContactCardChangesResponse is the arguments object returned by
// "ContactCard/changes".
type ContactCardChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}

// ContactCardQuery represents a JMAP "ContactCard/query" call
// (RFC 9610 §4.5, RFC 8620 §5.5).
type ContactCardQuery struct {
	CallID    string                    `json:"-"`
	AccountID string                    `json:"accountId"`
	Filter    *ContactCardFilter        `json:"filter,omitempty"`
	Sort      []ContactCardSort         `json:"sort,omitempty"`
	Position  int                       `json:"position,omitempty"`
	Limit     int                       `json:"limit,omitempty"`
	response  *ContactCardQueryResponse `json:"-"`
}

func (c *ContactCardQuery) Name() string { return "ContactCard/query" }
func (c *ContactCardQuery) ID() string   { return c.CallID }
func (c *ContactCardQuery) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/query result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (c *ContactCardQuery) Response() *ContactCardQueryResponse { return c.response }

// ContactCardQueryResponse is the arguments object returned by
// "ContactCard/query".
type ContactCardQueryResponse struct {
	AccountID           string   `json:"accountId"`
	QueryState          string   `json:"queryState"`
	CanCalculateChanges bool     `json:"canCalculateChanges"`
	Position            int      `json:"position"`
	IDs                 []string `json:"ids"`
	Total               *int     `json:"total,omitempty"`
	Limit               *int     `json:"limit,omitempty"`
}

// ContactCardFilter represents the filter argument to ContactCard/query
// (RFC 9610 §4.5).
type ContactCardFilter struct {
	InAddressBook string `json:"inAddressBook,omitempty"`
	UID           string `json:"uid,omitempty"`
	HasMember     string `json:"hasMember,omitempty"`
	Kind          string `json:"kind,omitempty"`

	CreatedBefore string `json:"createdBefore,omitempty"`
	CreatedAfter  string `json:"createdAfter,omitempty"`
	UpdatedBefore string `json:"updatedBefore,omitempty"`
	UpdatedAfter  string `json:"updatedAfter,omitempty"`

	Text          string `json:"text,omitempty"`
	Name          string `json:"name,omitempty"`
	NameGiven     string `json:"name/given,omitempty"`
	NameSurname   string `json:"name/surname,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Organization  string `json:"organization,omitempty"`
	Email         string `json:"email,omitempty"`
	Phone         string `json:"phone,omitempty"`
	OnlineService string `json:"onlineService,omitempty"`
	Address       string `json:"address,omitempty"`
	Note          string `json:"note,omitempty"`

	// Compound logic
	Operator   string               `json:"operator,omitempty"` // "AND", "OR", "NOT"
	Conditions []*ContactCardFilter `json:"conditions,omitempty"`
}

// ContactCardSort describes a sort criterion for ContactCard/query.
type ContactCardSort struct {
	Property    string `json:"property"` // "created", "updated", "name/given", "name/surname"
	IsAscending bool   `json:"isAscending,omitempty"`
}

// ContactCardQueryChanges represents a JMAP "ContactCard/queryChanges" call
// (RFC 9610 §4.6, RFC 8620 §5.6).
type ContactCardQueryChanges struct {
	CallID          string                           `json:"-"`
	AccountID       string                           `json:"accountId"`
	Filter          *ContactCardFilter               `json:"filter,omitempty"`
	Sort            []ContactCardSort                `json:"sort,omitempty"`
	SinceQueryState string                           `json:"sinceQueryState"`
	MaxChanges      *int                             `json:"maxChanges,omitempty"`
	UpToID          string                           `json:"upToId,omitempty"`
	CalculateTotal  bool                             `json:"calculateTotal,omitempty"`
	response        *ContactCardQueryChangesResponse `json:"-"`
}

func (c *ContactCardQueryChanges) Name() string { return "ContactCard/queryChanges" }
func (c *ContactCardQueryChanges) ID() string   { return c.CallID }
func (c *ContactCardQueryChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/queryChanges result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (c *ContactCardQueryChanges) Response() *ContactCardQueryChangesResponse { return c.response }

// ContactCardQueryChangesResponse is the arguments object returned by
// "ContactCard/queryChanges".
type ContactCardQueryChangesResponse struct {
	AccountID     string      `json:"accountId"`
	OldQueryState string      `json:"oldQueryState"`
	NewQueryState string      `json:"newQueryState"`
	Removed       []string    `json:"removed"`
	Added         []AddedItem `json:"added"`
	Total         *int        `json:"total,omitempty"`
}

// AddedItem represents a single entry in the "added" array of a
// queryChanges response (RFC 8620 §5.6).
type AddedItem struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
}

// ContactCardSet represents a JMAP "ContactCard/set" call (RFC 9610 §4.4).
type ContactCardSet struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`
	IfInState string `json:"ifInState,omitempty"`

	Create  map[string]*ContactCardCreate `json:"create,omitempty"`
	Update  map[string]*ContactCardPatch  `json:"update,omitempty"`
	Destroy []string                      `json:"destroy,omitempty"`

	response *ContactCardSetResponse `json:"-"`
}

func (c *ContactCardSet) Name() string { return "ContactCard/set" }
func (c *ContactCardSet) ID() string   { return c.CallID }
func (c *ContactCardSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/set result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (c *ContactCardSet) Response() *ContactCardSetResponse { return c.response }

// ContactCardCreate describes a new ContactCard to create. It embeds
// the common JSContact properties alongside the JMAP-specific
// AddressBookIDs field.
type ContactCardCreate struct {
	AddressBookIDs map[string]bool `json:"addressBookIds"`
	UID            string          `json:"uid,omitempty"`
	Kind           string          `json:"kind,omitempty"`

	Name      *Name                `json:"name,omitempty"`
	Nicknames map[string]*Nickname `json:"nicknames,omitempty"`

	Emails         map[string]*EmailAddress  `json:"emails,omitempty"`
	Phones         map[string]*Phone         `json:"phones,omitempty"`
	OnlineServices map[string]*OnlineService `json:"onlineServices,omitempty"`
	Addresses      map[string]*Address       `json:"addresses,omitempty"`

	Organizations map[string]*Organization `json:"organizations,omitempty"`
	Titles        map[string]*Title        `json:"titles,omitempty"`
	Notes         map[string]*Note         `json:"notes,omitempty"`
	Keywords      map[string]bool          `json:"keywords,omitempty"`
}

// ContactCardPatch represents the update object for a single ContactCard in
// a ContactCard/set call, serialized via [patch.Marshal].
//
// Each scalar field uses [patch.Value] for three-state semantics: absent
// (omitted from the patch), set (a concrete value), or null (explicit
// deletion). Map fields use [patch.Map] and can be wrapped with
// [patch.Partial] for per-entry patch paths.
type ContactCardPatch struct {
	// JMAP-specific
	AddressBookIDs patch.Map[bool] `json:"addressBookIds,omitempty"`

	// JSContact metadata
	Kind patch.Value[string] `json:"kind"`

	// Names
	Name      patch.Value[*Name]   `json:"name"`
	Nicknames patch.Map[*Nickname] `json:"nicknames,omitempty"`

	// Contact info
	Emails         patch.Map[*EmailAddress]  `json:"emails,omitempty"`
	Phones         patch.Map[*Phone]         `json:"phones,omitempty"`
	OnlineServices patch.Map[*OnlineService] `json:"onlineServices,omitempty"`

	// Addresses
	Addresses patch.Map[*Address] `json:"addresses,omitempty"`

	// Organizations and titles
	Organizations patch.Map[*Organization] `json:"organizations,omitempty"`
	Titles        patch.Map[*Title]        `json:"titles,omitempty"`

	// Other
	Notes    patch.Map[*Note] `json:"notes,omitempty"`
	Keywords patch.Map[bool]  `json:"keywords,omitempty"`
}

// MarshalJSON serializes the patch using [patch.Marshal].
func (p *ContactCardPatch) MarshalJSON() ([]byte, error) {
	return patch.Marshal(p)
}

// ContactCardSetResponse is the arguments object returned by
// "ContactCard/set".
type ContactCardSetResponse struct {
	AccountID    string                            `json:"accountId"`
	OldState     string                            `json:"oldState"`
	NewState     string                            `json:"newState"`
	Created      map[string]*ContactCardSetCreated `json:"created,omitempty"`
	Updated      map[string]*ContactCardSetUpdated `json:"updated,omitempty"`
	Destroyed    []string                          `json:"destroyed,omitempty"`
	NotCreated   map[string]*SetError              `json:"notCreated,omitempty"`
	NotUpdated   map[string]*SetError              `json:"notUpdated,omitempty"`
	NotDestroyed map[string]*SetError              `json:"notDestroyed,omitempty"`
}

// ContactCardSetCreated contains the server-assigned properties returned for
// a newly created ContactCard.
type ContactCardSetCreated struct {
	ID string `json:"id"`
}

// ContactCardSetUpdated may contain any server-side changes beyond what was
// requested. Typically empty.
type ContactCardSetUpdated struct{}

// ContactCardCopy represents a JMAP "ContactCard/copy" call
// (RFC 9610 §4.7, RFC 8620 §5.4).
type ContactCardCopy struct {
	CallID        string `json:"-"`
	FromAccountID string `json:"fromAccountId"`
	AccountID     string `json:"accountId"`

	IfFromInState string `json:"ifFromInState,omitempty"`
	IfInState     string `json:"ifInState,omitempty"`

	// Create maps client-assigned creation IDs to objects. Each object must
	// contain an "id" property with the ID of the card in the source account.
	Create map[string]*ContactCardCopyItem `json:"create"`

	// OnSuccessDestroyOriginal, if true, destroys the original card in the
	// source account after a successful copy.
	OnSuccessDestroyOriginal bool `json:"onSuccessDestroyOriginal,omitempty"`

	// DestroyFromIfInState is passed as "ifInState" to the implicit /set
	// call that destroys the original, if OnSuccessDestroyOriginal is true.
	DestroyFromIfInState string `json:"destroyFromIfInState,omitempty"`

	response *ContactCardCopyResponse `json:"-"`
}

func (c *ContactCardCopy) Name() string { return "ContactCard/copy" }
func (c *ContactCardCopy) ID() string   { return c.CallID }
func (c *ContactCardCopy) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &c.response)
}

// Response returns the decoded ContactCard/copy result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (c *ContactCardCopy) Response() *ContactCardCopyResponse { return c.response }

// ContactCardCopyItem describes a single card to copy. ID is the card's ID
// in the source account. AddressBookIDs can be set to place the copy into
// specific address books in the destination account.
type ContactCardCopyItem struct {
	ID             string          `json:"id"`
	AddressBookIDs map[string]bool `json:"addressBookIds,omitempty"`
}

// ContactCardCopyResponse is the arguments object returned by
// "ContactCard/copy".
type ContactCardCopyResponse struct {
	FromAccountID string                            `json:"fromAccountId"`
	AccountID     string                            `json:"accountId"`
	OldState      string                            `json:"oldState"`
	NewState      string                            `json:"newState"`
	Created       map[string]*ContactCardSetCreated `json:"created,omitempty"`
	NotCreated    map[string]*SetError              `json:"notCreated,omitempty"`
}
