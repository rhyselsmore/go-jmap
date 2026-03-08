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

### Update emails with Email/set

`EmailPatch` uses the [`protocol/patch`](../protocol/patch) package to control
serialization per field. By default a map field replaces the whole property on
the server. Wrap it with `patch.Partial` to expand individual entries into JMAP
patch paths instead.

```go
import "github.com/rhyselsmore/go-jmap/protocol/patch"

// Mark a single keyword without touching others (partial patch path).
kw := patch.Partial(patch.Map[bool]{})
kw.Set("$seen", true)

// Move to a new mailbox by replacing mailboxIds entirely (replace mode).
mb := patch.Map[bool]{}
mb.Set(newMailboxID, true)

req := jmap.NewRequest(core.Capability, mail.Capability)
req.Add(&mail.EmailSet{
    AccountID: accountID,
    Update: map[string]*mail.EmailPatch{
        emailID: {
            Keywords:   kw,
            MailboxIDs: mb,
        },
    },
})
```

To delete an entry from a map (set it to JSON null), use `Null`:

```go
kw.Null("$seen") // removes $seen without touching other keywords
```

Scalar fields like `Subject` use `patch.Value[T]`, which is a three-state type:
absent (omitted), set (a value), or null (explicit server-side deletion). Use
`patch.Set` to provide a value, or `patch.Null` to clear it:

```go
subj := "Re: hello"

req.Add(&mail.EmailSet{
    AccountID: accountID,
    Update: map[string]*mail.EmailPatch{
        // Update the subject.
        emailID: {Subject: patch.Set(subj)},
        // Clear the subject (set to null on the server).
        otherID: {Subject: patch.Null[string]()},
        // Leave subject unchanged (zero Value[T] is absent — field omitted).
        thirdID: {Keywords: kw},
    },
})
```

See the [top-level README](../README.md) for a full end-to-end example.
