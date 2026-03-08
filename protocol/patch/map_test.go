package patch

import "testing"

func TestMapIsPartial(t *testing.T) {
	tests := []struct {
		name    string
		m       Map[bool]
		partial bool
	}{
		{"new map is not partial", Map[bool]{}, false},
		{"make map is not partial", make(Map[bool]), false},
		{"Partial map is partial", Partial(Map[bool]{}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.IsPartial(); got != tt.partial {
				t.Errorf("IsPartial() = %v, want %v", got, tt.partial)
			}
		})
	}
}

func TestMapSet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"simple key", "key", "value"},
		{"JMAP keyword", "$seen", "x"},
		{"empty value", "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(Map[string])
			m.Set(tt.key, tt.value)
			got, ok := m[tt.key].Get()
			if !ok {
				t.Fatal("Set() entry Get() ok = false, want true")
			}
			if got != tt.value {
				t.Errorf("Set() entry Get() = %q, want %q", got, tt.value)
			}
		})
	}
}

func TestMapNull(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"simple key", "key"},
		{"JMAP keyword", "$seen"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := make(Map[string])
			m.Null(tt.key)
			if !m[tt.key].IsNull() {
				t.Error("Null() entry IsNull() = false, want true")
			}
			if m[tt.key].IsAbsent() {
				t.Error("Null() entry IsAbsent() = true, want false")
			}
		})
	}
}

func TestPartialReturnsSameMap(t *testing.T) {
	m := make(Map[bool])
	m.Set("a", true)
	p := Partial(m)
	p.Set("b", false)
	if _, ok := m["b"]; !ok {
		t.Error("Partial() should return the same map, not a copy")
	}
}
