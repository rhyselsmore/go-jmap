package jmap

import "testing"

func TestRef(t *testing.T) {
	inv := &testInvocation{name: "Mailbox/query", id: "c1"}
	ref := Ref(inv, "/ids/*")

	if ref.ResultOf != "c1" {
		t.Errorf("ResultOf = %q, want %q", ref.ResultOf, "c1")
	}
	if ref.Name != "Mailbox/query" {
		t.Errorf("Name = %q, want %q", ref.Name, "Mailbox/query")
	}
	if ref.Path != "/ids/*" {
		t.Errorf("Path = %q, want %q", ref.Path, "/ids/*")
	}
}
