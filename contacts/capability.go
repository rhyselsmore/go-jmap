package contacts

import "github.com/rhyselsmore/go-jmap"

// Capability is the JMAP capability identifier for contacts support, as
// defined in RFC 9610. Its presence in a Session indicates support for the
// AddressBook and ContactCard data types and associated API methods.
const Capability jmap.Capability = "urn:ietf:params:jmap:contacts"

// Capabilities is the server-level capability object for contacts. Per
// RFC 9610, this is an empty object; all contact-specific constraints are
// defined at the account level in [AccountCapabilities].
type Capabilities struct{}

// AccountCapabilities contains the server constraints and permissions for a
// contact-enabled account, as defined in RFC 9610 §2.
type AccountCapabilities struct {
	// MaxAddressBooksPerCard is the maximum number of AddressBooks that can
	// be assigned to a single ContactCard. Must be >= 1, or nil for no limit.
	MaxAddressBooksPerCard *int `json:"maxAddressBooksPerCard"`

	// MayCreateAddressBook indicates whether the user may create an
	// AddressBook in this account.
	MayCreateAddressBook *bool `json:"mayCreateAddressBook"`
}

// GetCapabilities decodes the server-level contacts capability from the
// Session. For contacts, this is an empty object; it primarily serves as a
// presence check to confirm the server supports the contacts capability.
func GetCapabilities(s jmap.Session) (*Capabilities, error) {
	return jmap.GetCapabilities[*Capabilities](s, Capability)
}

// GetAccountCapabilities decodes the account-level contacts capability from
// an Account, returning the server's constraints and permissions for contact
// operations on that account.
func GetAccountCapabilities(a jmap.Account) (*AccountCapabilities, error) {
	return jmap.GetAccountCapabilities[*AccountCapabilities](a, Capability)
}
