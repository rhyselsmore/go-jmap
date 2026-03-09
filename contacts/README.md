# contacts

Package `contacts` provides JMAP Contacts types and method call
implementations, as defined in [RFC 9610](https://www.rfc-editor.org/rfc/rfc9610).

Contact data uses the JSContact Card format
([RFC 9553](https://www.rfc-editor.org/rfc/rfc9553)).

## Implemented methods

| Method | Type | Description |
|---|---|---|
| `AddressBook/get` | `AddressBookGet` | Fetch AddressBook objects by ID |
| `AddressBook/changes` | `AddressBookChanges` | Fetch AddressBook IDs changed since a given state |
| `AddressBook/set` | `AddressBookSet` | Create, update, or destroy AddressBook objects |
| `ContactCard/get` | `ContactCardGet` | Fetch ContactCard objects by ID or result reference |
| `ContactCard/changes` | `ContactCardChanges` | Fetch ContactCard IDs changed since a given state |
| `ContactCard/query` | `ContactCardQuery` | Query ContactCards with optional filter and sort |
| `ContactCard/queryChanges` | `ContactCardQueryChanges` | Diff a ContactCard/query result set since a given query state |
| `ContactCard/set` | `ContactCardSet` | Create, update, or destroy ContactCard objects |
| `ContactCard/copy` | `ContactCardCopy` | Copy ContactCards between accounts |

## Usage

### Inspect account capabilities

```go
import "github.com/rhyselsmore/go-jmap/contacts"

session, _ := client.GetSession(ctx)
accountID := session.PrimaryAccounts[contacts.Capability]

caps, err := contacts.GetAccountCapabilities(session.Accounts[accountID])
if err != nil {
    log.Fatal(err)
}
fmt.Printf("May create address books: %v\n", *caps.MayCreateAddressBook)
```

### Fetch all address books

```go
import (
    "github.com/rhyselsmore/go-jmap"
    "github.com/rhyselsmore/go-jmap/core"
    "github.com/rhyselsmore/go-jmap/contacts"
)

req := jmap.NewRequest(core.Capability, contacts.Capability)

get := &contacts.AddressBookGet{AccountID: accountID}
req.Add(get)

if _, err := client.Do(ctx, req); err != nil {
    log.Fatal(err)
}

for _, ab := range get.Response().List {
    fmt.Printf("%s (default=%v)\n", ab.Name, ab.IsDefault)
}
```

### Create an address book

```go
req := jmap.NewRequest(core.Capability, contacts.Capability)
req.Add(&contacts.AddressBookSet{
    AccountID: accountID,
    Create: map[string]*contacts.AddressBookCreate{
        "new0": {Name: "Work Contacts"},
    },
})
```

### Query and fetch contacts

```go
q := &contacts.ContactCardQuery{
    AccountID: accountID,
    Filter:    &contacts.ContactCardFilter{Email: "alice@example.com"},
}
req.Add(q)

get := &contacts.ContactCardGet{
    AccountID: accountID,
    IDRef:     jmap.Ref(q, "/ids/*"),
}
req.Add(get)
```

### Create a contact

```go
req.Add(&contacts.ContactCardSet{
    AccountID: accountID,
    Create: map[string]*contacts.ContactCardCreate{
        "new0": {
            AddressBookIDs: map[string]bool{addressBookID: true},
            Name: &contacts.Name{
                Components: []contacts.NameComponent{
                    {Kind: "given", Value: "Alice"},
                    {Kind: "surname", Value: "Smith"},
                },
            },
            Emails: map[string]*contacts.EmailAddress{
                "e1": {Address: "alice@example.com", Contexts: map[string]bool{"work": true}},
            },
            Phones: map[string]*contacts.Phone{
                "p1": {Number: "+1-555-0100", Features: map[string]bool{"cell": true}},
            },
        },
    },
})
```

### Sync contacts with ContactCard/changes

```go
ch := &contacts.ContactCardChanges{
    AccountID:  accountID,
    SinceState: cachedState,
}
req.Add(ch)

get := &contacts.ContactCardGet{
    AccountID: accountID,
    IDRef:     jmap.Ref(ch, "/created"),
}
req.Add(get)
```

### Copy contacts between accounts

```go
req.Add(&contacts.ContactCardCopy{
    FromAccountID: sourceAccountID,
    AccountID:     destAccountID,
    Create: map[string]*contacts.ContactCardCopyItem{
        "copy0": {
            ID:             cardID,
            AddressBookIDs: map[string]bool{destAddressBookID: true},
        },
    },
})
```

See the [top-level README](../README.md) for a full end-to-end example.
