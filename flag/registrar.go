package flag

import (
	"reflect"
	"strings"
	"unsafe"
)

// Registrar abstracts objects that have flags that need to be registered to
// a [Registry].
//
// Registrars are queried as part of the [Register] operation, which allows for
// recursive and conditional registration of custom flag types.
type Registrar interface {
	RegisterFlags(fs *Registry)
}

// Register adds all flags associated to v into fs.
//
// If v implements [Registrar] directly, it will be registered immediately.
//
// If v does not implement [Registrar], but is an iterable type like a slice,
// map, or structured object -- each visible field will recursively be inspected
// for whether it implements [Registrar], and any discovered fields will be
// registered.
//
// Note:
// If an object implements [Registrar] and contains fields of other types
// that may be [Registrar] types, the object is responsible for manually
// registering those fields.
func Register(registry *Registry, v any) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	register(registry, rv, rt)
}

type tags struct {
	ignore bool
}

func parseTags(tag reflect.StructTag) tags {
	var result tags
	value, ok := tag.Lookup("flag")
	if !ok {
		return result
	}
	for part := range strings.SplitSeq(value, ",") {
		switch part {
		case "ignore", "-":
			result.ignore = true
		}
	}
	return result
}

func register(registry *Registry, rv reflect.Value, rt reflect.Type) {
	if !rv.CanInterface() {
		return
	}
	for {
		if registrar, ok := rv.Interface().(Registrar); ok {
			if id, ok := instanceID(rv); ok {
				if _, seen := registry.visited[id]; seen {
					return
				}
				registry.visited[id] = struct{}{}
			}
			registrar.RegisterFlags(registry)
			return
		}
		kind := rt.Kind()
		if kind != reflect.Pointer && kind != reflect.Interface {
			break
		}
		rv = rv.Elem()
		rt = rv.Type()
	}

	switch rt.Kind() {
	case reflect.Struct:
		registerStruct(registry, rv, rt)
	case reflect.Slice, reflect.Array:
		registerSlice(registry, rv)
	case reflect.Map:
		registerMap(registry, rv)
	}
}

// instanceID returns the pointer identity of the [Registrar] value rv, peeling
// interface wrappers first. It reports false for non-pointer registrars, which
// have no shared identity and are always registered.
func instanceID(rv reflect.Value) (unsafe.Pointer, bool) {
	for rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Pointer && !rv.IsNil() {
		return rv.UnsafePointer(), true
	}
	return nil, false
}

func registerStruct(registry *Registry, rv reflect.Value, rt reflect.Type) {
	for _, field := range reflect.VisibleFields(rt) {
		tag := parseTags(field.Tag)
		if tag.ignore {
			continue
		}
		fieldV := rv.FieldByIndex(field.Index)
		fieldT := field.Type
		register(registry, fieldV, fieldT)
	}
}

func registerSlice(registry *Registry, rv reflect.Value) {
	length := rv.Len()
	for i := range length {
		fieldV := rv.Index(i)
		fieldT := fieldV.Type()
		register(registry, fieldV, fieldT)
	}
}

func registerMap(registry *Registry, rv reflect.Value) {
	if rv.IsNil() {
		return
	}
	for _, key := range rv.MapKeys() {
		register(registry, key, key.Type())
		value := rv.MapIndex(key)
		register(registry, value, value.Type())
	}
}
