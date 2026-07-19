package arg

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/bitwizeshift/go-cli/internal/format/csvfield"
)

// Unmarshaler is implemented by types that decode themselves from the raw bytes
// of an arg value.
type Unmarshaler interface {
	// UnmarshalArg decodes data into the receiver, returning an error if the
	// input is not valid for the type.
	UnmarshalArg(data []byte) error
}

var (
	// ErrInvalidTarget indicates the value passed to [Unmarshal] was not a
	// non-nil pointer that can be written to.
	ErrInvalidTarget = errors.New("target must be a non-nil pointer")

	// ErrUnsupportedType indicates the element type addressed by the target is
	// not one that [Unmarshal] knows how to decode.
	ErrUnsupportedType = errors.New("unsupported type")
)

// durationType identifies [time.Duration] so it may be decoded via
// [time.ParseDuration] rather than as a plain integer.
var durationType = reflect.TypeFor[time.Duration]()

// Unmarshal decodes data into out.
//
// The element addressed by out is decoded by the first of these that applies:
// an [Unmarshaler] implementation, an [encoding.TextUnmarshaler] implementation,
// or -- for a string, boolean, integer, floating-point, or [time.Duration]
// element -- built-in parsing:
//
//   - Integers are parsed with the base inferred from the input (e.g. "0x" for
//     hexadecimal, "0b" for binary, "0o" for octal, otherwise decimal).
//   - Floating-point values are parsed as [strconv.ParseFloat] does.
//   - [time.Duration] is parsed with [time.ParseDuration].
//   - Strings are taken verbatim.
//   - Booleans accept only "true" or "false".
//   - Slices parse the input as comma-separated values, decoding each field by
//     the rules above for the element type.
//
// This function will return [ErrInvalidTarget] if out is not a writable pointer
// (nor a self-decoding value), [ErrUnsupportedType] if the element type is not
// supported, or any errors from the underlying API otherwise.
func Unmarshal(out any, data []byte) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer {
		if ok, err := tryUnmarshaler(out, data); ok {
			return err
		}
		return fmt.Errorf("unmarshal: %w", ErrInvalidTarget)
	}
	if rv.IsNil() {
		return fmt.Errorf("unmarshal: %w", ErrInvalidTarget)
	}

	val, err := decodeElem(rv.Type().Elem(), data)
	if err != nil {
		return err
	}
	rv.Elem().Set(val)
	return nil
}

// decodeElem decodes data into a fresh value of type t, following any pointer
// indirection as [Unmarshal] does, and returns the settable result.
func decodeElem(t reflect.Type, data []byte) (reflect.Value, error) {
	ptr := reflect.New(elemType(t))
	if err := decode(ptr, data); err != nil {
		return reflect.Value{}, err
	}
	out := reflect.New(t).Elem()
	assign(out, ptr.Elem())
	return out, nil
}

// decode fills the value addressed by ptr, a non-nil pointer, from data. It
// prefers an [Unmarshaler] or [encoding.TextUnmarshaler] implementation on ptr,
// falling back to decoding the addressed value by its kind.
func decode(ptr reflect.Value, data []byte) error {
	if ok, err := tryUnmarshaler(ptr.Interface(), data); ok {
		return err
	}
	return unmarshalInto(ptr.Elem(), string(data))
}

// tryUnmarshaler decodes data into v when v implements [Unmarshaler] or
// [encoding.TextUnmarshaler], reporting whether either applied.
func tryUnmarshaler(v any, data []byte) (bool, error) {
	switch u := v.(type) {
	case Unmarshaler:
		return true, u.UnmarshalArg(data)
	case encoding.TextUnmarshaler:
		return true, u.UnmarshalText(data)
	}
	return false, nil
}

// elemType returns the non-pointer element type addressed by pointer type t,
// following through any number of levels of indirection.
func elemType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// assign stores decoded into dst, allocating any nil pointers encountered while
// following dst's chain of indirection.
func assign(dst, decoded reflect.Value) {
	for dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}
	dst.Set(decoded)
}

// unmarshalInto decodes s into the settable value v according to its type.
func unmarshalInto(v reflect.Value, s string) error {
	if v.Type() == durationType {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		v.SetInt(int64(d))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 0, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 0, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		v.SetFloat(f)
	case reflect.Slice:
		parts, err := csvfield.Split(s)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
		slice := reflect.MakeSlice(v.Type(), len(parts), len(parts))
		for i, part := range parts {
			elem, err := decodeElem(v.Type().Elem(), []byte(part))
			if err != nil {
				return err
			}
			slice.Index(i).Set(elem)
		}
		v.Set(slice)
	default:
		return fmt.Errorf("unmarshal: %w: %s", ErrUnsupportedType, v.Type())
	}
	return nil
}
