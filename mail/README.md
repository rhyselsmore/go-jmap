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
| `Mailbox/set` | `MailboxSet` | Create, update, or destroy Mailbox objects |
| `Mailbox/changes` | `MailboxChanges` | Fetch mailbox IDs changed since a given state |
| `Mailbox/queryChanges` | `MailboxQueryChanges` | Diff a Mailbox/query result set since a given query state |
| `Thread/get` | `ThreadGet` | Fetch Thread objects by ID or result reference |
| `Thread/changes` | `ThreadChanges` | Fetch thread IDs changed since a given state |
| `VacationResponse/get` | `VacationResponseGet` | Fetch the singleton VacationResponse object |
| `VacationResponse/set` | `VacationResponseSet` | Update the singleton VacationResponse object |
| `Email/changes` | `EmailChanges` | Fetch email IDs changed since a given state |
| `Email/queryChanges` | `EmailQueryChanges` | Diff an Email/query result set since a given query state |
| `SearchSnippet/get` | `SearchSnippetGet` | Fetch highlighted search snippets for emails |

## Unimplemented methods

The following RFC 8621 methods are not yet implemented:

| Method | RFC 8621 Section |
|---|---|
| `Email/copy` | §4.7 |
| `Email/import` | §4.8 |

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

### Create and update mailboxes with Mailbox/set

```go
import "github.com/rhyselsmore/go-jmap/protocol/patch"

role := mail.RoleArchive
req := jmap.NewRequest(core.Capability, mail.Capability)
req.Add(&mail.MailboxSet{
    AccountID: accountID,
    Create: map[string]*mail.MailboxCreate{
        "new0": {Name: "Archive", Role: &role},
    },
    Update: map[string]*mail.MailboxPatch{
        existingID: {Name: patch.Set("Renamed Folder")},
    },
    Destroy: []string{oldMailboxID},
})
```

### Sync mailboxes with Mailbox/changes

Use the state string from a previous `Mailbox/get` to fetch only what changed,
then feed the created and updated IDs into a `Mailbox/get` — all in one request.

```go
ch := &mail.MailboxChanges{
    AccountID:  accountID,
    SinceState: cachedState,
}
req.Add(ch)

get := &mail.MailboxGet{
    AccountID: accountID,
    IDRef:     jmap.Ref(ch, "/created"),
}
req.Add(get)
```

After execution, remove `ch.Response().Destroyed` IDs from your local cache and
store `ch.Response().NewState` for next time.

### Sync emails with Email/changes

The same pattern works for emails:

```go
ch := &mail.EmailChanges{
    AccountID:  accountID,
    SinceState: cachedEmailState,
}
req.Add(ch)

get := &mail.EmailGet{
    AccountID: accountID,
    IDRef:     jmap.Ref(ch, "/created"),
}
req.Add(get)
```

### Fetch threads

Threads group related emails. Use `Thread/get` with IDs from an email's
`ThreadID` field, or via a result reference:

```go
threads := &mail.ThreadGet{
    AccountID: accountID,
    IDs:       []string{threadID},
}
req.Add(threads)
```

Each `Thread` in the response contains an `EmailIDs` slice sorted by
`receivedAt` (oldest first).

### Search with highlighted snippets

Pair `Email/query` with `SearchSnippet/get` to get search results with
highlighted matching terms in a single round trip:

```go
q := &mail.EmailQuery{
    AccountID: accountID,
    Filter:    &mail.EmailFilter{Text: "quarterly report"},
    Limit:     20,
}
req.Add(q)

snippets := &mail.SearchSnippetGet{
    AccountID:  accountID,
    Filter:     q.Filter,
    EmailIDRef: jmap.Ref(q, "/ids/*"),
}
req.Add(snippets)
```

The returned `SearchSnippet` objects contain `Subject` and `Preview` fields
with matching terms wrapped in HTML `<mark>` tags.

### Manage vacation auto-replies

`VacationResponse` is a singleton — there is exactly one per account, always
with the ID `"singleton"`. It requires the `VacationResponseCapability` in the
`using` array.

```go
req := jmap.NewRequest(core.Capability, mail.Capability, mail.VacationResponseCapability)

// Fetch current settings.
get := &mail.VacationResponseGet{AccountID: accountID}
req.Add(get)

// Enable auto-replies.
subject := "Out of office"
body := "I'm away until Monday."
req.Add(&mail.VacationResponseSet{
    AccountID: accountID,
    Update: map[string]*mail.VacationResponsePatch{
        "singleton": {
            IsEnabled: ptrBool(true),
            Subject:   &subject,
            TextBody:  &body,
        },
    },
})
```

See the [top-level README](../README.md) for a full end-to-end example.
