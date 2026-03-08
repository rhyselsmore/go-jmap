package jmap

import (
	"encoding/json"
	"testing"
)

func TestNewRequest(t *testing.T) {
	r := NewRequest("urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail")
	if len(r.Using) != 2 {
		t.Fatalf("Using length = %d, want 2", len(r.Using))
	}
	if r.Using[0] != "urn:ietf:params:jmap:core" {
		t.Errorf("Using[0] = %q, want %q", r.Using[0], "urn:ietf:params:jmap:core")
	}
	if len(r.MethodCalls) != 0 {
		t.Errorf("MethodCalls length = %d, want 0", len(r.MethodCalls))
	}
}

func TestRequestAdd(t *testing.T) {
	t.Run("auto-generates IDs", func(t *testing.T) {
		r := NewRequest()
		r.Add(&testInvocation{name: "Mailbox/get"})
		r.Add(&testInvocation{name: "Email/query"})

		if len(r.ids) != 2 {
			t.Fatalf("ids length = %d, want 2", len(r.ids))
		}
		if r.ids[0].id != "c1" {
			t.Errorf("ids[0].id = %q, want %q", r.ids[0].id, "c1")
		}
		if r.ids[1].id != "c2" {
			t.Errorf("ids[1].id = %q, want %q", r.ids[1].id, "c2")
		}
	})

	t.Run("uses provided ID", func(t *testing.T) {
		r := NewRequest()
		r.Add(&testInvocation{name: "Mailbox/get", id: "myid"})

		if r.ids[0].id != "myid" {
			t.Errorf("ids[0].id = %q, want %q", r.ids[0].id, "myid")
		}
	})

	t.Run("deduplicates IDs", func(t *testing.T) {
		r := NewRequest()
		r.Add(&testInvocation{name: "Mailbox/get", id: "dup"})
		r.Add(&testInvocation{name: "Email/get", id: "dup"})

		if r.ids[1].id != "dup.1" {
			t.Errorf("deduplicated id = %q, want %q", r.ids[1].id, "dup.1")
		}
	})

	t.Run("triple dedup", func(t *testing.T) {
		r := NewRequest()
		r.Add(&testInvocation{name: "A", id: "x"})
		r.Add(&testInvocation{name: "B", id: "x"})
		r.Add(&testInvocation{name: "C", id: "x"})

		if r.ids[2].id != "x.2" {
			t.Errorf("third deduplicated id = %q, want %q", r.ids[2].id, "x.2")
		}
	})
}

func TestRequestMarshalJSON(t *testing.T) {
	r := NewRequest("urn:ietf:params:jmap:core")
	r.Add(&testInvocation{name: "Mailbox/get", id: "c1"})

	got, err := r.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	using, ok := parsed["using"].([]any)
	if !ok || len(using) != 1 || using[0] != "urn:ietf:params:jmap:core" {
		t.Errorf("using = %v", parsed["using"])
	}

	calls, ok := parsed["methodCalls"].([]any)
	if !ok || len(calls) != 1 {
		t.Fatalf("methodCalls length = %v, want 1", len(calls))
	}
	triple, ok := calls[0].([]any)
	if !ok || len(triple) != 3 {
		t.Fatalf("triple length = %v, want 3", len(triple))
	}
	if triple[0] != "Mailbox/get" {
		t.Errorf("method name = %v, want Mailbox/get", triple[0])
	}
	if triple[2] != "c1" {
		t.Errorf("call id = %v, want c1", triple[2])
	}
}

func TestRequestLookup(t *testing.T) {
	r := NewRequest()
	inv := &testInvocation{name: "Mailbox/get", id: "c1"}
	r.Add(inv)

	t.Run("found", func(t *testing.T) {
		got, ok := r.lookup("c1")
		if !ok {
			t.Fatal("lookup() ok = false, want true")
		}
		if got != inv {
			t.Error("lookup() returned wrong invocation")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := r.lookup("missing")
		if ok {
			t.Error("lookup() ok = true, want false")
		}
	})
}
