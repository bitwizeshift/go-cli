package ask

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Unmarshal parses text into the value pointed to by v.
//
// If v implements [encoding.TextUnmarshaler] it is used directly. A
// *[time.Duration] is parsed with [time.ParseDuration]; otherwise the pointed-to
// value is set by reflecting on its [reflect.Kind], covering strings, booleans,
// and the sized integer, unsigned, and floating-point types.
//
// It returns any parse error for malformed input. It panics when v is not a
// non-nil pointer to a supported type, which is a programming error rather than
// a user-input error.
func Unmarshal(text string, v any) error {
	if u, ok := v.(encoding.TextUnmarshaler); ok {
		return u.UnmarshalText([]byte(text))
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic(fmt.Sprintf("ask: Unmarshal into non-pointer %T", v))
	}

	elem := rv.Elem()
	if elem.Type() == reflect.TypeFor[time.Duration]() {
		duration, err := time.ParseDuration(text)
		if err != nil {
			return err
		}
		elem.SetInt(int64(duration))
		return nil
	}

	switch elem.Kind() {
	case reflect.String:
		elem.SetString(text)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}
		elem.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(text, 0, elem.Type().Bits())
		if err != nil {
			return err
		}
		elem.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(text, 0, elem.Type().Bits())
		if err != nil {
			return err
		}
		elem.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(text, elem.Type().Bits())
		if err != nil {
			return err
		}
		elem.SetFloat(parsed)
	default:
		panic(fmt.Sprintf("ask: unsupported target type %T", v))
	}
	return nil
}
