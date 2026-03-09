package mail

import (
	"encoding/json"

	"github.com/rhyselsmore/go-jmap"
)

// SearchSnippet represents a JMAP SearchSnippet object as defined in
// RFC 8621 §5. It contains fragments of the email subject and body with
// matching search terms highlighted using HTML <mark> tags.
type SearchSnippet struct {
	EmailID string  `json:"emailId"`
	Subject *string `json:"subject"` // null if no matches in subject
	Preview *string `json:"preview"` // null if no matches in body
}

// SearchSnippetGet represents a JMAP "SearchSnippet/get" call (RFC 8621 §5.1).
//
// Unlike a standard /get, this method requires a Filter (the same as used in
// Email/query) so the server knows which search terms to highlight. It also
// uses EmailIDs (or EmailIDRef) instead of the usual IDs field.
type SearchSnippetGet struct {
	CallID     string                `json:"-"`
	AccountID  string                `json:"accountId"`
	Filter     *EmailFilter          `json:"filter,omitempty"`
	EmailIDs   []string              `json:"emailIds,omitempty"`
	EmailIDRef *jmap.ResultReference `json:"#emailIds,omitempty"`
	response   *SearchSnippetGetResponse `json:"-"`
}

func (s *SearchSnippetGet) Name() string { return "SearchSnippet/get" }
func (s *SearchSnippetGet) ID() string   { return s.CallID }
func (s *SearchSnippetGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &s.response)
}

// Response returns the decoded SearchSnippet/get result. It is only populated
// after the request has been executed via [jmap.Client.Do].
func (s *SearchSnippetGet) Response() *SearchSnippetGetResponse { return s.response }

// SearchSnippetGetResponse is the arguments object returned by
// "SearchSnippet/get".
type SearchSnippetGetResponse struct {
	AccountID string          `json:"accountId"`
	List      []SearchSnippet `json:"list"`
	NotFound  []string        `json:"notFound,omitempty"`
}
