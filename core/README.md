# core

Package `core` provides the JMAP Core capability identifier and server
capability types, as defined in [RFC 8620](https://www.rfc-editor.org/rfc/rfc8620).

The core capability (`urn:ietf:params:jmap:core`) is **required** in every
JMAP request. It describes the server's fundamental limits — request sizes,
concurrency, and object counts — that clients must respect.

## Usage

Pass `core.Capability` when building a request, and use `core.GetCapabilities`
to inspect the server's limits from the session.

```go
import (
    "github.com/rhyselsmore/go-jmap"
    "github.com/rhyselsmore/go-jmap/core"
)

session, _ := client.GetSession(ctx)

caps, err := core.GetCapabilities(session)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Max calls per request: %d\n", caps.MaxCallsInRequest)
fmt.Printf("Max upload size:       %d bytes\n", caps.MaxSizeUpload)

// Always include core.Capability in every request.
req := jmap.NewRequest(core.Capability, /* other capabilities */)
```

## Types

| Type | Description |
|---|---|
| `Capability` | Constant URI for the core capability |
| `Capabilities` | Server-level limits decoded from the Session |

See the [top-level README](../README.md) for a full end-to-end example.
