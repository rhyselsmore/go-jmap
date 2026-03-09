package contacts

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap/protocol/patch"
)

// AddressBook represents a JMAP AddressBook object as defined in RFC 9610 §3.
type AddressBook struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  *string            `json:"description,omitempty"`
	SortOrder    int                `json:"sortOrder"`
	IsDefault    bool               `json:"isDefault"`
	IsSubscribed bool               `json:"isSubscribed"`
	MyRights     *AddressBookRights `json:"myRights,omitempty"`
	ShareWith    map[string]*AddressBookRights `json:"shareWith,omitempty"`
}

// AddressBookRights describes the permissions a user has on an AddressBook,
// as defined in RFC 9610 §3.
type AddressBookRights struct {
	MayRead   bool `json:"mayRead"`
	MayWrite  bool `json:"mayWrite"`
	MayShare  bool `json:"mayShare"`
	MayDelete bool `json:"mayDelete"`
}

// AddressBookGet represents a JMAP "AddressBook/get" call (RFC 9610 §3.2).
type AddressBookGet struct {
	CallID     string                      `json:"-"`
	AccountID  string                      `json:"accountId"`
	IDs        []string                    `json:"ids,omitempty"`
	Properties []string                    `json:"properties,omitempty"`
	response   *AddressBookGetResponse     `json:"-"`
}

func (a *AddressBookGet) Name() string { return "AddressBook/get" }
func (a *AddressBookGet) ID() string   { return a.CallID }
func (a *AddressBookGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &a.response)
}

// Response returns the decoded AddressBook/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (a *AddressBookGet) Response() *AddressBookGetResponse { return a.response }

// AddressBookGetResponse is the arguments object returned by "AddressBook/get".
type AddressBookGetResponse struct {
	AccountID string        `json:"accountId"`
	State     string        `json:"state"`
	List      []AddressBook `json:"list"`
	NotFound  []string      `json:"notFound,omitempty"`
}

// AddressBookChanges represents a JMAP "AddressBook/changes" call
// (RFC 9610 §3.3, RFC 8620 §5.2).
type AddressBookChanges struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// SinceState is the state string from a previous AddressBook/get response.
	SinceState string `json:"sinceState"`

	// MaxChanges limits the number of IDs returned. If more changes exist,
	// HasMoreChanges will be true in the response.
	MaxChanges *int `json:"maxChanges,omitempty"`

	response *AddressBookChangesResponse `json:"-"`
}

func (a *AddressBookChanges) Name() string { return "AddressBook/changes" }
func (a *AddressBookChanges) ID() string   { return a.CallID }
func (a *AddressBookChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &a.response)
}

// Response returns the decoded AddressBook/changes result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (a *AddressBookChanges) Response() *AddressBookChangesResponse { return a.response }

// AddressBookChangesResponse is the arguments object returned by
// "AddressBook/changes".
type AddressBookChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}

// AddressBookSet represents a JMAP "AddressBook/set" call (RFC 9610 §3.4).
type AddressBookSet struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// IfInState, if set, must match the current state string from
	// AddressBook/get. If it does not match, the server rejects the call
	// with a stateMismatch error.
	IfInState string `json:"ifInState,omitempty"`

	Create  map[string]*AddressBookCreate `json:"create,omitempty"`
	Update  map[string]*AddressBookPatch  `json:"update,omitempty"`
	Destroy []string                      `json:"destroy,omitempty"`

	// OnDestroyRemoveContents controls what happens to ContactCards in a
	// destroyed AddressBook. If true, cards that exist only in the destroyed
	// AddressBook are also destroyed. If false (the default), the server
	// rejects the destroy when cards would be orphaned.
	OnDestroyRemoveContents bool `json:"onDestroyRemoveContents,omitempty"`

	response *AddressBookSetResponse `json:"-"`
}

func (a *AddressBookSet) Name() string { return "AddressBook/set" }
func (a *AddressBookSet) ID() string   { return a.CallID }
func (a *AddressBookSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &a.response)
}

// Response returns the decoded AddressBook/set result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (a *AddressBookSet) Response() *AddressBookSetResponse { return a.response }

// AddressBookCreate describes a new AddressBook to create in an
// AddressBook/set call.
type AddressBookCreate struct {
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	SortOrder    int     `json:"sortOrder,omitempty"`
	IsSubscribed bool    `json:"isSubscribed,omitempty"`
}

// AddressBookPatch represents the update object for a single AddressBook in
// an AddressBook/set call, serialized via [patch.Marshal].
//
// Each field uses [patch.Value] for three-state semantics: absent (omitted
// from the patch), set (a concrete value), or null (explicit deletion).
type AddressBookPatch struct {
	Name         patch.Value[string] `json:"name"`
	Description  patch.Value[string] `json:"description"`
	SortOrder    patch.Value[int]    `json:"sortOrder"`
	IsSubscribed patch.Value[bool]   `json:"isSubscribed"`
}

// MarshalJSON serializes the patch using [patch.Marshal].
func (p *AddressBookPatch) MarshalJSON() ([]byte, error) {
	return patch.Marshal(p)
}

// AddressBookSetResponse is the arguments object returned by
// "AddressBook/set".
type AddressBookSetResponse struct {
	AccountID    string                               `json:"accountId"`
	OldState     string                               `json:"oldState"`
	NewState     string                               `json:"newState"`
	Created      map[string]*AddressBookSetCreated    `json:"created,omitempty"`
	Updated      map[string]*AddressBookSetUpdated    `json:"updated,omitempty"`
	Destroyed    []string                             `json:"destroyed,omitempty"`
	NotCreated   map[string]*SetError                 `json:"notCreated,omitempty"`
	NotUpdated   map[string]*SetError                 `json:"notUpdated,omitempty"`
	NotDestroyed map[string]*SetError                 `json:"notDestroyed,omitempty"`
}

// AddressBookSetCreated contains the server-assigned properties returned for
// a newly created AddressBook.
type AddressBookSetCreated struct {
	ID string `json:"id"`
}

// AddressBookSetUpdated may contain any server-side changes beyond what was
// requested. Typically empty.
type AddressBookSetUpdated struct{}

// SetError represents an error for a single create, update, or destroy
// operation within a /set response, as defined in RFC 8620 §5.3.
type SetError struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}
