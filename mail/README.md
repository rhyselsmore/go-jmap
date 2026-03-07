# mail

Package `mail` provides JMAP Mail types and method call implementations, as
defined in [RFC 8621](https://www.rfc-editor.org/rfc/rfc8621).

## Implemented methods

| Method | Type | Description |
|---|---|---|
| `Mailbox/query` | `MailboxQuery` | Query mailboxes with optional filter and sort |
| `Mailbox/get` | `MailboxGet` | Fetch full Mailbox objects by ID or result reference |
| `Email/query` | `EmailQuery` | Query emails with optional filter and sort |
| `Email/get` | `EmailGet` | Fetch full Email objects by ID or result reference |
| `Email/set` | `EmailSet` | Create, update, or destroy Email objects |

## Usage

### Inspect account capabilities

```go
import "github.com/rhyselsmore/go-jmap/mail"

session, _ := client.GetSession(ctx)
accountID := session.PrimaryAccounts[mail.Capability]

caps, err := mail.GetAccountCapabilities(session.Accounts[accountID])
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Max mailbox depth: %v\n", caps.MaxMailboxDepth)
```

### Query and fetch mailboxes in one round trip

Use `IDRef` with `jmap.Ref` to pass the results of a query directly into a
get call — the server resolves the reference without a second round trip.

```go
import (
    "github.com/rhyselsmore/go-jmap"
    "github.com/rhyselsmore/go-jmap/core"
    "github.com/rhyselsmore/go-jmap/mail"
)

req := jmap.NewRequest(core.Capability, mail.Capability)

q1 := &mail.MailboxQuery{AccountID: accountID}
req.Add(q1)

q2 := &mail.MailboxGet{
    AccountID: accountID,
    IDRef:     jmap.Ref(q1, "/ids/*"),
}
req.Add(q2)

if _, err := client.Do(ctx, req); err != nil {
    log.Fatal(err)
}

for _, mb := range q2.Response().List {
    fmt.Printf("%s (%s)\n", mb.Name, mb.ID)
}
```

### Query and fetch emails

```go
q1 := &mail.EmailQuery{
    AccountID: accountID,
    Filter:    &mail.EmailFilter{HasKeyword: "$unseen"},
    Limit:     50,
}
req.Add(q1)

q2 := &mail.EmailGet{
    AccountID:           accountID,
    IDRef:               jmap.Ref(q1, "/ids/*"),
    FetchTextBodyValues: true,
}
req.Add(q2)
```

See the [top-level README](../README.md) for a full end-to-end example.
