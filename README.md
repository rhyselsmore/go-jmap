# go-jmap

> **Work in progress — not yet usable.**
> The API is unstable, incomplete, and will change without notice. Do not use this in production.

A Go client library for the [JMAP](https://jmap.io) protocol ([RFC 8620](https://www.rfc-editor.org/rfc/rfc8620)).

## Status

This library is in early development. Currently implemented:

- JMAP session discovery and caching
- Request/response envelope encoding (RFC 8620 §3)
- Bearer token authentication
- `Mailbox/query` & `Mailbox/get` method calls

Much of the JMAP Core and JMAP Mail surface area is not yet covered.

## Example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rhyselsmore/go-jmap"
	"github.com/rhyselsmore/go-jmap/core"
	"github.com/rhyselsmore/go-jmap/mail"
)

func main() {
	// Create a client with bearer token authentication.
	client, err := jmap.NewClient(
		jmap.WithBearerTokenAuthentication(os.Getenv("FASTMAIL_API_TOKEN")),
		jmap.WithStaticResolver("https://api.fastmail.com"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Fetch the JMAP session. This is cached for subsequent calls.
	session, err := client.GetSession(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Look up the primary mail account and inspect capabilities.
	accountId := session.PrimaryAccounts[mail.Capability]
	caps, _ := mail.GetAccountCapabilities(session.Accounts[accountId])
	fmt.Printf("Max mailbox depth: %v\n", caps.MaxMailboxDepth)

	// Build a request with two dependent calls: query all mailboxes,
	// then fetch them using a result reference — resolved server-side
	// in a single round trip.
	req := jmap.NewRequest(core.Capability, mail.Capability)

	q1 := &mail.MailboxQuery{
		AccountID: accountId,
	}
	req.Add(q1)

	q2 := &mail.MailboxGet{
		AccountID: accountId,
		IDRef:     jmap.Ref(q1, "/ids/*"),
	}
	req.Add(q2)

	if _, err = client.Do(context.Background(), req); err != nil {
		log.Fatal(err)
	}

	// Responses are available directly on the invocation objects.
	for _, mb := range q2.Response().List {
		fmt.Printf("Mailbox: %s (%s)\n", mb.Name, mb.ID)
	}
}
```

## Requirements

- Go 1.25+

## License

MIT
