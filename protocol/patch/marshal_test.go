package patch

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
)

// testRawMap is a map whose values don't implement json.Marshaler,
// used to exercise the non-Marshaler branch in Marshal.
type testRawMap map[string]string

func (testRawMap) patchMap() {}

func (m testRawMap) IsPartial() bool {
	_, ok := m[partialMapKey]
	return ok
}

func TestMarshal(t *testing.T) {
	type simple struct {
		Name  Value[string] `json:"name"`
		Count Value[int]    `json:"count"`
	}

	type withMap struct {
		Keywords Map[bool] `json:"keywords"`
	}

	type withPatchTag struct {
		Keywords Map[bool] `json:"keywords" patch:"mailboxIds"`
	}

	type withRegular struct {
		Name  string     `json:"name"`
		Count Value[int] `json:"count"`
	}

	type withJSONDash struct {
		Name   Value[string] `json:"name"`
		Hidden Value[string] `json:"-"`
	}

	type withUnexported struct {
		Name     Value[string] `json:"name"`
		internal string
	}

	type noTag struct {
		FieldName Value[string]
	}

	type commaOnly struct {
		Field Value[string] `json:",omitempty"`
	}

	type withRaw struct {
		Tags testRawMap `json:"tags"`
	}

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "all values set",
			input: simple{Name: Set("inbox"), Count: Set(5)},
			want:  `{"count":5,"name":"inbox"}`,
		},
		{
			name:  "absent value omitted",
			input: simple{Name: Set("inbox")},
			want:  `{"name":"inbox"}`,
		},
		{
			name:  "null value emitted",
			input: simple{Name: Null[string](), Count: Set(1)},
			want:  `{"count":1,"name":null}`,
		},
		{
			name:  "all absent produces empty object",
			input: simple{},
			want:  `{}`,
		},
		{
			name: "map replace mode",
			input: func() withMap {
				s := withMap{Keywords: make(Map[bool])}
				s.Keywords.Set("$seen", true)
				s.Keywords.Set("$flagged", false)
				return s
			}(),
			want: `{"keywords":{"$flagged":false,"$seen":true}}`,
		},
		{
			name: "map replace mode with null entry",
			input: func() withMap {
				s := withMap{Keywords: make(Map[bool])}
				s.Keywords.Set("$seen", true)
				s.Keywords.Null("$draft")
				return s
			}(),
			want: `{"keywords":{"$draft":null,"$seen":true}}`,
		},
		{
			name: "map partial mode",
			input: func() withMap {
				s := withMap{Keywords: Partial(Map[bool]{})}
				s.Keywords.Set("$seen", true)
				s.Keywords.Null("$flagged")
				return s
			}(),
			want: `{"keywords/$flagged":null,"keywords/$seen":true}`,
		},
		{
			name: "map partial with patch tag prefix",
			input: func() withPatchTag {
				s := withPatchTag{Keywords: Partial(Map[bool]{})}
				s.Keywords.Set("$draft", true)
				return s
			}(),
			want: `{"mailboxIds/$draft":true}`,
		},
		{
			name:  "nil map omitted",
			input: withMap{},
			want:  `{}`,
		},
		{
			name:  "empty map replace mode omitted",
			input: withMap{Keywords: make(Map[bool])},
			want:  `{}`,
		},
		{
			name: "empty partial map produces no entries",
			input: func() withMap {
				return withMap{Keywords: Partial(Map[bool]{})}
			}(),
			want: `{}`,
		},
		{
			name:  "regular field zero value omitted",
			input: withRegular{Count: Set(1)},
			want:  `{"count":1}`,
		},
		{
			name:  "regular field non-zero emitted",
			input: withRegular{Name: "inbox", Count: Set(1)},
			want:  `{"count":1,"name":"inbox"}`,
		},
		{
			name:  "json dash tag skipped",
			input: withJSONDash{Name: Set("visible"), Hidden: Set("secret")},
			want:  `{"name":"visible"}`,
		},
		{
			name:  "unexported fields skipped",
			input: withUnexported{Name: Set("test"), internal: "hidden"},
			want:  `{"name":"test"}`,
		},
		{
			name:  "pointer input dereferenced",
			input: &simple{Name: Set("ptr")},
			want:  `{"name":"ptr"}`,
		},
		{
			name:  "no json tag uses field name",
			input: noTag{FieldName: Set("val")},
			want:  `{"FieldName":"val"}`,
		},
		{
			name:  "json tag with only options uses field name",
			input: commaOnly{Field: Set("val")},
			want:  `{"Field":"val"}`,
		},
		{
			name:  "raw map replace mode (non-Marshaler values)",
			input: withRaw{Tags: testRawMap{"color": "red"}},
			want:  `{"tags":{"color":"red"}}`,
		},
		{
			name: "raw map partial mode (non-Marshaler values)",
			input: func() withRaw {
				m := testRawMap{"color": "red"}
				m[partialMapKey] = ""
				return withRaw{Tags: m}
			}(),
			want: `{"tags/color":"red"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			assertJSONEqual(t, got, tt.want)
		})
	}
}

func TestMarshalErrors(t *testing.T) {
	type badValueStruct struct {
		Val Value[chan int] `json:"val"`
	}

	type badMapStruct struct {
		Items Map[chan int] `json:"items"`
	}

	tests := []struct {
		name  string
		input any
	}{
		{
			name:  "value marshal error",
			input: badValueStruct{Val: Set(make(chan int))},
		},
		{
			name: "map replace mode marshal error",
			input: func() badMapStruct {
				s := badMapStruct{Items: make(Map[chan int])}
				s.Items.Set("a", make(chan int))
				return s
			}(),
		},
		{
			name: "map partial mode marshal error",
			input: func() badMapStruct {
				s := badMapStruct{Items: Partial(Map[chan int]{})}
				s.Items.Set("a", make(chan int))
				return s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Marshal(tt.input)
			if err == nil {
				t.Fatal("Marshal() expected error")
			}
		})
	}
}

func TestMarshalFieldMetaCache(t *testing.T) {
	type cached struct {
		A Value[string] `json:"a"`
	}

	first, err := Marshal(cached{A: Set("first")})
	if err != nil {
		t.Fatalf("first Marshal() error: %v", err)
	}

	second, err := Marshal(cached{A: Set("second")})
	if err != nil {
		t.Fatalf("second Marshal() error: %v", err)
	}

	assertJSONEqual(t, first, `{"a":"first"}`)
	assertJSONEqual(t, second, `{"a":"second"}`)
}

func TestMarshalConcurrent(t *testing.T) {
	type concStruct struct {
		X Value[int] `json:"x"`
	}

	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Marshal(concStruct{X: Set(i)})
			if err != nil {
				t.Errorf("Marshal() error: %v", err)
			}
		}()
	}
	wg.Wait()
}

func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotVal, wantVal any
	if err := json.Unmarshal(got, &gotVal); err != nil {
		t.Fatalf("unmarshal got: %v\nraw: %s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &wantVal); err != nil {
		t.Fatalf("unmarshal want: %v\nraw: %s", err, want)
	}
	if !reflect.DeepEqual(gotVal, wantVal) {
		t.Errorf("JSON mismatch\ngot:  %s\nwant: %s", got, want)
	}
}
