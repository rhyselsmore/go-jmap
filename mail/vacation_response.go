package mail

import (
	"encoding/json"
	"time"

	"github.com/rhyselsmore/go-jmap"
)

// VacationResponseCapability is the JMAP capability identifier for vacation
// response support, as defined in RFC 8621 §6. This is separate from the
// mail capability and must be included in the "using" array when making
// VacationResponse method calls.
const VacationResponseCapability jmap.Capability = "urn:ietf:params:jmap:vacationresponse"

// VacationResponse represents the JMAP VacationResponse object as defined in
// RFC 8621 §6. There is exactly one VacationResponse per account, always with
// the ID "singleton".
type VacationResponse struct {
	ID        string     `json:"id"`                  // always "singleton"
	IsEnabled bool       `json:"isEnabled"`           // whether auto-replies are active
	FromDate  *time.Time `json:"fromDate,omitempty"`  // start of vacation period (UTC)
	ToDate    *time.Time `json:"toDate,omitempty"`    // end of vacation period (UTC)
	Subject   *string    `json:"subject,omitempty"`   // subject line for auto-replies
	TextBody  *string    `json:"textBody,omitempty"`  // plain text body
	HTMLBody  *string    `json:"htmlBody,omitempty"`  // HTML body, if supported
}

// VacationResponseGet represents a JMAP "VacationResponse/get" call
// (RFC 8621 §6.1). It follows the standard /get pattern defined in
// RFC 8620 §5.1. Because VacationResponse is a singleton, IDs should
// be ["singleton"] or omitted.
type VacationResponseGet struct {
	CallID     string                        `json:"-"`
	AccountID  string                        `json:"accountId"`
	IDs        []string                      `json:"ids,omitempty"`
	Properties []string                      `json:"properties,omitempty"`
	response   *VacationResponseGetResponse  `json:"-"`
}

func (v *VacationResponseGet) Name() string { return "VacationResponse/get" }
func (v *VacationResponseGet) ID() string   { return v.CallID }
func (v *VacationResponseGet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &v.response)
}

// Response returns the decoded VacationResponse/get result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (v *VacationResponseGet) Response() *VacationResponseGetResponse { return v.response }

// VacationResponseGetResponse is the arguments object returned by
// "VacationResponse/get".
type VacationResponseGetResponse struct {
	AccountID string             `json:"accountId"`
	State     string             `json:"state"`
	List      []VacationResponse `json:"list"`
	NotFound  []string           `json:"notFound,omitempty"`
}

// VacationResponseSet represents a JMAP "VacationResponse/set" call
// (RFC 8621 §6.2). It follows the standard /set pattern defined in
// RFC 8620 §5.3. Because VacationResponse is a singleton, the only valid
// key in Update is "singleton".
type VacationResponseSet struct {
	CallID    string `json:"-"`
	AccountID string `json:"accountId"`

	// IfInState, if set, must match the current state string from
	// VacationResponse/get. If it does not match, the server rejects the
	// call with a stateMismatch error.
	IfInState string `json:"ifInState,omitempty"`

	// Update maps object IDs to patch objects. For VacationResponse the
	// only valid key is "singleton".
	Update map[string]*VacationResponsePatch `json:"update,omitempty"`

	response *VacationResponseSetResponse `json:"-"`
}

func (v *VacationResponseSet) Name() string { return "VacationResponse/set" }
func (v *VacationResponseSet) ID() string   { return v.CallID }
func (v *VacationResponseSet) DecodeResponse(b json.RawMessage) error {
	return json.Unmarshal(b, &v.response)
}

// Response returns the decoded VacationResponse/set result. It is only
// populated after the request has been executed via [jmap.Client.Do].
func (v *VacationResponseSet) Response() *VacationResponseSetResponse { return v.response }

// VacationResponsePatch is the update object for VacationResponse/set.
// Each field is a pointer so that absent fields (nil) are omitted from
// the JSON patch, while present fields (including zero values) are sent.
type VacationResponsePatch struct {
	IsEnabled *bool       `json:"isEnabled,omitempty"`
	FromDate  *time.Time  `json:"fromDate,omitempty"`
	ToDate    *time.Time  `json:"toDate,omitempty"`
	Subject   *string     `json:"subject,omitempty"`
	TextBody  *string     `json:"textBody,omitempty"`
	HTMLBody  *string     `json:"htmlBody,omitempty"`
}

// VacationResponseSetResponse is the arguments object returned by
// "VacationResponse/set".
type VacationResponseSetResponse struct {
	AccountID  string                                    `json:"accountId"`
	OldState   string                                    `json:"oldState"`
	NewState   string                                    `json:"newState"`
	Updated    map[string]*VacationResponseSetUpdated    `json:"updated,omitempty"`
	NotUpdated map[string]*SetError                      `json:"notUpdated,omitempty"`
}

// VacationResponseSetUpdated may contain any server-side changes beyond what
// was requested. Typically empty.
type VacationResponseSetUpdated struct{}
