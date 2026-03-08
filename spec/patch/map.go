package patch

// partialMapKey is a sentinel key used to store the partial mode flag inside
// the map itself. The value is intentionally bizarre to avoid colliding with
// real JMAP IDs, which are defined in RFC 8620 as ASCII-only strings.
const partialMapKey = "_not-expecting-this-to-collide__jmap_patch_is_partial-💩"

// Map is a map of JMAP patch values, where each entry is a three-state value
// that can be set, null (delete), or absent (omit from patch).
//
// By default a Map serializes as a complete property replacement. Wrap it
// with [Partial] to switch to partial mode, where each entry is expanded into
// a JMAP patch path (e.g. "keywords/$seen": true) rather than replacing the
// whole property.
//
// Map is not safe for concurrent use; callers are responsible for
// synchronizing access.
type Map[T any] map[string]Value[T]

// patchMap implements the mapPatch marker interface so that patch.Marshal
// can identify Map fields via reflection.
func (Map[T]) patchMap() {}

// IsPartial reports whether the map is in partial mode.
func (m Map[T]) IsPartial() bool {
	_, ok := m[partialMapKey]
	return ok
}

// Set marks the entry at k to be set to v in the patch.
func (m Map[T]) Set(k string, v T) {
	m[k] = Value[T]{value: &v, isSet: true}
}

// Null marks the entry at k to be deleted in the patch.
func (m Map[T]) Null(k string) {
	m[k] = Value[T]{isSet: true}
}

// Partial returns m in partial mode. In partial mode, entries are expanded
// into JMAP patch paths during serialization (e.g. "keywords/$seen": true)
// rather than replacing the whole property. The default is replace mode.
func Partial[T any](m Map[T]) Map[T] {
	m[partialMapKey] = Value[T]{}
	return m
}
