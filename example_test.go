package jmap_test

import (
	"encoding/json"
	"fmt"

	jmap "github.com/rhyselsmore/go-jmap"
)

func ExampleNewBearerTokenAuthenticator() {
	authn, err := jmap.NewBearerTokenAuthenticator("my-token")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(authn != nil)
	// Output: true
}

func ExampleGetCapabilities() {
	sess := jmap.Session{
		Capabilities: map[jmap.Capability]json.RawMessage{
			"urn:ietf:params:jmap:core": json.RawMessage(`{"maxSizeUpload":50000000}`),
		},
	}

	type CoreCapability struct {
		MaxSizeUpload int `json:"maxSizeUpload"`
	}

	cap, err := jmap.GetCapabilities[CoreCapability](sess, "urn:ietf:params:jmap:core")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cap.MaxSizeUpload)
	// Output: 50000000
}

func ExampleNewStaticResolver() {
	r, err := jmap.NewStaticResolver("https://api.fastmail.com")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(r != nil)
	// Output: true
}

func ExampleNewRequest() {
	req := jmap.NewRequest("urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail")
	data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))
	// Output: {"using":["urn:ietf:params:jmap:core","urn:ietf:params:jmap:mail"],"methodCalls":[]}
}

func ExampleRef() {
	// Ref creates a back-reference to use the output of one method call
	// as input to another within the same request.
	inv := &exampleInvocation{name: "Mailbox/query", id: "c1"}
	ref := jmap.Ref(inv, "/ids/*")
	fmt.Println(ref.ResultOf, ref.Name, ref.Path)
	// Output: c1 Mailbox/query /ids/*
}

// exampleInvocation implements jmap.Invocation for examples.
type exampleInvocation struct {
	name string
	id   string
}

func (e *exampleInvocation) Name() string                          { return e.name }
func (e *exampleInvocation) ID() string                            { return e.id }
func (e *exampleInvocation) DecodeResponse(_ json.RawMessage) error { return nil }
