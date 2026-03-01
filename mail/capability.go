package mail

import (
	"github.com/rhyselsmore/go-jmap"
)

// Capability is the JMAP capability identifier for mail support, as defined
// in RFC 8621. Its presence in a Session indicates support for the Mailbox,
// Thread, Email, and SearchSnippet data types and associated API methods.
const Capability jmap.Capability = "urn:ietf:params:jmap:mail"

// Capabilities is the server-level capability object for mail. Per RFC 8621,
// this is an empty object; all mail-specific constraints are defined at the
// account level in AccountCapabilities.
type Capabilities struct{}

// AccountCapabilities contains the server constraints and permissions
// for a mail-enabled account, as defined in RFC 8621 Section 2.
type AccountCapabilities struct {
	// MaxMailboxesPerEmail is the maximum number of Mailboxes that can be
	// assigned to a single Email object. Must be >= 1, or nil for no limit.
	MaxMailboxesPerEmail *int `json:"maxMailboxesPerEmail"`

	// MaxMailboxDepth is the maximum depth of the Mailbox hierarchy
	// (one more than the maximum number of ancestors a Mailbox may have),
	// or nil for no limit.
	MaxMailboxDepth *int `json:"maxMailboxDepth"`

	// MaxSizeMailboxName is the maximum length, in UTF-8 octets, allowed
	// for the name of a Mailbox. Must be at least 100.
	MaxSizeMailboxName int `json:"maxSizeMailboxName"`

	// MaxSizeAttachmentsPerEmail is the maximum total size of attachments,
	// in octets, allowed for a single Email object. This is the sum of
	// unencoded attachment sizes, matching what users see on disk.
	MaxSizeAttachmentsPerEmail int `json:"maxSizeAttachmentsPerEmail"`

	// EmailQuerySortOptions lists all values the server supports for the
	// "property" field of the Comparator object in an Email/query sort.
	// May include properties the client does not recognise.
	EmailQuerySortOptions []string `json:"emailQuerySortOptions"`

	// MayCreateTopLevelMailbox indicates whether the user may create a
	// Mailbox at the top level of the hierarchy (i.e., with no parent).
	MayCreateTopLevelMailbox bool `json:"mayCreateTopLevelMailbox"`
}

// GetCapabilities decodes the server-level mail capability from the Session.
// For mail, this is an empty object; it primarily serves as a presence check
// to confirm the server supports the mail capability.
func GetCapabilities(s jmap.Session) (*Capabilities, error) {
	return jmap.GetCapabilities[*Capabilities](s, Capability)
}

// GetAccountCapabilities decodes the account-level mail capability from an
// Account, returning the server's constraints and permissions for mail
// operations on that account.
func GetAccountCapabilities(a jmap.Account) (*AccountCapabilities, error) {
	return jmap.GetAccountCapabilities[*AccountCapabilities](a, Capability)
}
