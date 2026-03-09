package mail

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap"
)

// Thread represents a JMAP Thread object as defined in RFC 8621 §3.
// A thread groups related emails; the server determines the threading
// algorithm. EmailIDs are sorted by receivedAt date, oldest first.
type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

// ThreadGet represents a JMAP "Thread/get" call (RFC 8621 §3.1).
// It follows the standard /get pattern defined in RFC 8620 §5.1.
type ThreadGet struct {
	CallID     string                `json:"-"`
	AccountID  string                `json:"accountId"`
	IDs        []string              `json:"ids,omitempty"`
	IDRef      *jmap.ResultReference `json:"#ids,omitempty"`
	Properties []string              `json:"properties,omitempty"`
	response   *ThreadGetResponse    `json:"-"`
}

func (t *ThreadGet) Name() string { return "Thread/get" }
func (t *ThreadGet) ID() string   { return t.CallID }
func (t *ThreadGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &t.response)
}

// Response returns the decoded Thread/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (t *ThreadGet) Response() *ThreadGetResponse { return t.response }

// ThreadGetResponse is the arguments object returned by "Thread/get".
type ThreadGetResponse struct {
	AccountID string   `json:"accountId"`
	State     string   `json:"state"`
	List      []Thread `json:"list"`
	NotFound  []string `json:"notFound,omitempty"`
}

// ThreadChanges represents a JMAP "Thread/changes" call (RFC 8621 §3.2,
// RFC 8620 §5.2). It returns the IDs of threads that have been created,
// updated, or destroyed since the given state.
type ThreadChanges struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// SinceState is the state string from a previous Thread/get response.
	SinceState string `json:"sinceState"`

	// MaxChanges limits the number of IDs returned. If more changes exist,
	// HasMoreChanges will be true in the response.
	MaxChanges *int `json:"maxChanges,omitempty"`

	response *ThreadChangesResponse `json:"-"`
}

func (t *ThreadChanges) Name() string { return "Thread/changes" }
func (t *ThreadChanges) ID() string   { return t.CallID }
func (t *ThreadChanges) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &t.response)
}

// Response returns the decoded Thread/changes result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (t *ThreadChanges) Response() *ThreadChangesResponse { return t.response }

// ThreadChangesResponse is the arguments object returned by "Thread/changes".
type ThreadChangesResponse struct {
	AccountID      string   `json:"accountId"`
	OldState       string   `json:"oldState"`
	NewState       string   `json:"newState"`
	HasMoreChanges bool     `json:"hasMoreChanges"`
	Created        []string `json:"created"`
	Updated        []string `json:"updated"`
	Destroyed      []string `json:"destroyed"`
}
