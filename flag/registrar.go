package flag

import (
	"reflect"
	"strings"

	"github.com/spf13/pflag"
)

// Registrar abstracts objects that have flags that need to be registered to
// a [pflag.FlagSet].
//
// Registrars are queried as part of the [Register] operation, which allows for
// recursive and conditional registration of custom flag types.
type Registrar interface {
	RegisterFlags(fs *pflag.FlagSet)
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
func Register(fs *pflag.FlagSet, v any) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	register(fs, rv, rt)
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

func register(fs *pflag.FlagSet, rv reflect.Value, rt reflect.Type) {
	if !rv.CanInterface() {
		return
	}
	for {
		if registrar, ok := rv.Interface().(Registrar); ok {
			registrar.RegisterFlags(fs)
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
		registerStruct(fs, rv, rt)
	case reflect.Slice, reflect.Array:
		registerSlice(fs, rv)
	case reflect.Map:
		registerMap(fs, rv)
	}
}

func registerStruct(fs *pflag.FlagSet, rv reflect.Value, rt reflect.Type) {
	for _, field := range reflect.VisibleFields(rt) {
		tag := parseTags(field.Tag)
		if tag.ignore {
			continue
		}
		fieldV := rv.FieldByIndex(field.Index)
		fieldT := field.Type
		register(fs, fieldV, fieldT)
	}
}

func registerSlice(fs *pflag.FlagSet, rv reflect.Value) {
	length := rv.Len()
	for i := range length {
		fieldV := rv.Index(i)
		fieldT := fieldV.Type()
		register(fs, fieldV, fieldT)
	}
}

func registerMap(fs *pflag.FlagSet, rv reflect.Value) {
	if rv.IsNil() {
		return
	}
	for _, key := range rv.MapKeys() {
		register(fs, key, key.Type())
		value := rv.MapIndex(key)
		register(fs, value, value.Type())
	}
}
