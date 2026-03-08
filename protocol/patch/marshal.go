package patch

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// fieldMeta holds precomputed metadata for a single struct field.
type fieldMeta struct {
	index       int
	jsonName    string
	patchPrefix string // json field name used as path prefix in partial mode
	isMap       bool   // true if field is patch.Map[T]
	isValue     bool   // true if field is patch.Value[T]
}

var (
	cacheMu sync.RWMutex
	cache   = make(map[reflect.Type][]fieldMeta)
)

// mapPatch and valuePatch are marker interfaces used to identify
// patch.Map and patch.Value fields via reflection.
type mapPatch interface{ patchMap() }
type valuePatch interface{ patchValue() }

var (
	mapPatchType   = reflect.TypeOf((*mapPatch)(nil)).Elem()
	valuePatchType = reflect.TypeOf((*valuePatch)(nil)).Elem()
)

// Marshal serializes a patch struct to JSON. Each patch.Map field is
// serialized according to its own mode: in partial mode, entries are expanded
// into JMAP patch paths (e.g. "keywords/$seen": true); in replace mode, the
// map is serialized as a whole property replacement. patch.Value fields are
// emitted as direct property replacements if non-absent. Unexported fields
// and fields tagged `json:"-"` are skipped.
func Marshal(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	rt := rv.Type()

	fields := getFieldMeta(rt)
	out := make(map[string]any)

	for _, f := range fields {
		fv := rv.Field(f.index)

		if f.isMap {
			// Ask the map itself whether to expand into patch paths.
			partial := fv.MethodByName("IsPartial").Call(nil)[0].Bool()
			if partial {
				iter := fv.MapRange()
				for iter.Next() {
					k := iter.Key().String()
					if k == partialMapKey {
						continue
					}
					key := f.patchPrefix + "/" + k
					entry := iter.Value().Interface()
					if m, ok := entry.(json.Marshaler); ok {
						raw, err := m.MarshalJSON()
						if err != nil {
							return nil, fmt.Errorf("patch: marshaling %s: %w", key, err)
						}
						out[key] = json.RawMessage(raw)
					} else {
						out[key] = entry
					}
				}
			} else {
				// Replace mode: serialize the map wholesale, excluding sentinel.
				tmp := make(map[string]any)
				iter := fv.MapRange()
				for iter.Next() {
					k := iter.Key().String()
					if k == partialMapKey {
						continue
					}
					entry := iter.Value().Interface()
					if m, ok := entry.(json.Marshaler); ok {
						raw, err := m.MarshalJSON()
						if err != nil {
							return nil, fmt.Errorf("patch: marshaling %s/%s: %w", f.jsonName, k, err)
						}
						tmp[k] = json.RawMessage(raw)
					} else {
						tmp[k] = entry
					}
				}
				if len(tmp) > 0 {
					out[f.jsonName] = tmp
				}
			}
			continue
		}

		if f.isValue {
			absent := fv.MethodByName("IsAbsent").Call(nil)[0].Bool()
			if absent {
				continue
			}
			raw, err := fv.Interface().(json.Marshaler).MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("patch: marshaling %s: %w", f.jsonName, err)
			}
			out[f.jsonName] = json.RawMessage(raw)
			continue
		}

		// Regular field - omit zero values.
		if !fv.IsZero() {
			out[f.jsonName] = fv.Interface()
		}
	}

	return json.Marshal(out)
}

// getFieldMeta returns cached field metadata for t, computing it on first access.
func getFieldMeta(t reflect.Type) []fieldMeta {
	cacheMu.RLock()
	if fields, ok := cache[t]; ok {
		cacheMu.RUnlock()
		return fields
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-checked locking.
	if fields, ok := cache[t]; ok {
		return fields
	}

	fields := computeFieldMeta(t)
	cache[t] = fields
	return fields
}

// computeFieldMeta inspects t and builds field metadata from json struct tags.
// The patch path prefix for Map fields defaults to the json field name.
func computeFieldMeta(t reflect.Type) []fieldMeta {
	var fields []fieldMeta

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		jsonTag := sf.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = sf.Name
		}

		// patch tag overrides the path prefix for Map fields; defaults to jsonName.
		patchPrefix := sf.Tag.Get("patch")
		if patchPrefix == "" {
			patchPrefix = jsonName
		}

		fm := fieldMeta{
			index:       i,
			jsonName:    jsonName,
			patchPrefix: patchPrefix,
		}

		ft := sf.Type
		if ft.Implements(mapPatchType) || reflect.PointerTo(ft).Implements(mapPatchType) {
			fm.isMap = true
		} else if ft.Implements(valuePatchType) || reflect.PointerTo(ft).Implements(valuePatchType) {
			fm.isValue = true
		}

		fields = append(fields, fm)
	}

	return fields
}
