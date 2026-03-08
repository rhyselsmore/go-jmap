package patch

import "testing"

func TestValue(t *testing.T) {
	tests := []struct {
		name   string
		value  Value[string]
		absent bool
		null   bool
		get    string
		getOK  bool
		json   string
	}{
		{
			name:   "zero value is absent",
			value:  Value[string]{},
			absent: true,
			null:   false,
			get:    "",
			getOK:  false,
			json:   "null",
		},
		{
			name:   "Null is explicit null",
			value:  Null[string](),
			absent: false,
			null:   true,
			get:    "",
			getOK:  false,
			json:   "null",
		},
		{
			name:   "Set with non-empty string",
			value:  Set("hello"),
			absent: false,
			null:   false,
			get:    "hello",
			getOK:  true,
			json:   `"hello"`,
		},
		{
			name:   "Set with empty string",
			value:  Set(""),
			absent: false,
			null:   false,
			get:    "",
			getOK:  true,
			json:   `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsAbsent(); got != tt.absent {
				t.Errorf("IsAbsent() = %v, want %v", got, tt.absent)
			}
			if got := tt.value.IsNull(); got != tt.null {
				t.Errorf("IsNull() = %v, want %v", got, tt.null)
			}
			gotVal, gotOK := tt.value.Get()
			if gotOK != tt.getOK {
				t.Errorf("Get() ok = %v, want %v", gotOK, tt.getOK)
			}
			if gotVal != tt.get {
				t.Errorf("Get() = %q, want %q", gotVal, tt.get)
			}
			raw, err := tt.value.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error: %v", err)
			}
			if string(raw) != tt.json {
				t.Errorf("MarshalJSON() = %s, want %s", raw, tt.json)
			}
		})
	}
}

func TestValueInt(t *testing.T) {
	tests := []struct {
		name  string
		value Value[int]
		get   int
		getOK bool
		json  string
	}{
		{
			name:  "Set zero is still set",
			value: Set(0),
			get:   0,
			getOK: true,
			json:  "0",
		},
		{
			name:  "Set non-zero",
			value: Set(42),
			get:   42,
			getOK: true,
			json:  "42",
		},
		{
			name:  "Null int",
			value: Null[int](),
			get:   0,
			getOK: false,
			json:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.value.Get()
			if ok != tt.getOK {
				t.Errorf("Get() ok = %v, want %v", ok, tt.getOK)
			}
			if got != tt.get {
				t.Errorf("Get() = %d, want %d", got, tt.get)
			}
			raw, err := tt.value.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error: %v", err)
			}
			if string(raw) != tt.json {
				t.Errorf("MarshalJSON() = %s, want %s", raw, tt.json)
			}
		})
	}
}

func TestValueBool(t *testing.T) {
	v := Set(false)
	if v.IsAbsent() {
		t.Error("Set(false) should not be absent")
	}
	got, ok := v.Get()
	if !ok {
		t.Error("Set(false) Get() ok = false, want true")
	}
	if got != false {
		t.Error("Set(false) Get() = true, want false")
	}
}

func TestValueMarshalJSONError(t *testing.T) {
	v := Set(make(chan int))
	_, err := v.MarshalJSON()
	if err == nil {
		t.Fatal("MarshalJSON() expected error for unmarshalable type")
	}
}
