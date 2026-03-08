package patch

import "encoding/json"

// Value represents a three-state value in a JMAP patch operation.
// The zero value is absent, meaning the key will be omitted from the patch.
type Value[T any] struct {
	value *T
	isSet bool
}

func (Value[T]) patchValue() {}

// Set returns a Value that will serialize to the given value.
func Set[T any](v T) Value[T] {
	return Value[T]{value: &v, isSet: true}
}

// Null returns a Value that will serialize to JSON null,
// instructing the server to delete the corresponding entry.
func Null[T any]() Value[T] {
	return Value[T]{isSet: true}
}

// IsAbsent reports whether this value was not included in the patch.
// Absent values are omitted from the serialized patch entirely.
func (v Value[T]) IsAbsent() bool {
	return !v.isSet
}

// IsNull reports whether this value is an explicit null,
// meaning the corresponding server-side entry will be deleted.
func (v Value[T]) IsNull() bool {
	return v.isSet && v.value == nil
}

// Get returns the underlying value and whether it is non-null.
// If the value is absent or null, the returned T is the zero value.
func (v Value[T]) Get() (T, bool) {
	if v.value == nil {
		var zero T
		return zero, false
	}
	return *v.value, true
}

// MarshalJSON serializes the value as JSON null if explicitly null,
// or as the underlying value if set.
// Absent values should never be marshaled; callers are responsible
// for omitting absent Values from their containing map.
func (v Value[T]) MarshalJSON() ([]byte, error) {
	if v.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*v.value)
}
