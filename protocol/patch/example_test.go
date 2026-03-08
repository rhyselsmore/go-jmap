package patch_test

import (
	"fmt"

	"github.com/rhyselsmore/go-jmap/protocol/patch"
)

func ExampleSet() {
	v := patch.Set("hello")
	fmt.Println(v.IsAbsent())
	fmt.Println(v.IsNull())
	val, ok := v.Get()
	fmt.Println(val, ok)
	// Output:
	// false
	// false
	// hello true
}

func ExampleNull() {
	v := patch.Null[string]()
	fmt.Println(v.IsAbsent())
	fmt.Println(v.IsNull())
	_, ok := v.Get()
	fmt.Println("has value:", ok)
	// Output:
	// false
	// true
	// has value: false
}

func ExampleValue_absent() {
	var v patch.Value[string]
	fmt.Println("absent:", v.IsAbsent())
	fmt.Println("null:", v.IsNull())
	// Output:
	// absent: true
	// null: false
}

func ExampleMarshal() {
	type MailboxPatch struct {
		Name     patch.Value[string] `json:"name"`
		ParentID patch.Value[string] `json:"parentId"`
	}

	p := MailboxPatch{
		Name:     patch.Set("Inbox"),
		ParentID: patch.Null[string](),
	}

	data, err := patch.Marshal(p)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))
	// Output: {"name":"Inbox","parentId":null}
}

func ExampleMarshal_partialMap() {
	type MailboxPatch struct {
		Keywords patch.Map[bool] `json:"keywords"`
	}

	p := MailboxPatch{
		Keywords: patch.Partial(patch.Map[bool]{}),
	}
	p.Keywords.Set("$seen", true)
	p.Keywords.Null("$flagged")

	data, err := patch.Marshal(p)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(data))
	// Output: {"keywords/$flagged":null,"keywords/$seen":true}
}

func ExamplePartial() {
	m := patch.Partial(patch.Map[bool]{})
	m.Set("$seen", true)
	fmt.Println(m.IsPartial())
	// Output: true
}
