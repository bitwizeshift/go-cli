package arg

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/strcase"
	"github.com/spf13/pflag"
)

// typeToName converts the underlying object's value to a pflag "type" name
// by converting the Go reflect-API's type-identifier to a kebab-case. This
// includes the package the object comes from in the name, so `mips.OpCode`
// will become `mips-op-code`.
func typeToName(v any) string {
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	name := rt.String()
	parts := strings.Split(name, ".")
	for i := range parts {
		parts[i] = strcase.ToKebab(parts[i])
	}
	return strings.Join(parts, "-")
}

// Option configures an optional property of a flag constructed by [Flag].
type Option interface {
	apply(*config)
}

type option func(*config)

func (o option) apply(c *config) { o(c) }

// DefaultFunc is a function that can provide a flag default, only executed
// if a flag was not set during the CLI invocation.
type DefaultFunc func(ctx context.Context) (string, error)

// config holds the resolved options for a single flag registration.
type config struct {
	shorthand string
	usage     string
	hidden    bool
	required  bool
	typeName  func(any) string
	set       func(any, []byte) error

	// Callbacks invoked with the decoded value each time the flag is set.
	callbacks []reflect.Value

	// Occurrence limits.

	maxSet   bool // an occurrence option was applied
	maxCount int  // when maxSet: 0 means unlimited, n > 0 caps at n
	capped   bool // RepeatableUpTo was used, selecting ErrTooManyOccurrences

	// Fallbacks

	envs   []string
	custom []DefaultFunc

	// Shell completion.
	completer completerFunc
}

// newConfig builds a config from options, defaulting the type name to
// [typeToName] and the decoder to [Unmarshal].
func newConfig(options ...Option) *config {
	cfg := &config{typeName: typeToName, set: Unmarshal}
	for _, opt := range options {
		opt.apply(cfg)
	}
	return cfg
}

// Shorthand sets the single-character shorthand alias for the flag.
func Shorthand(short string) Option {
	return option(func(c *config) { c.shorthand = short })
}

// Usage sets the help string displayed for the flag.
func Usage(usage string) Option {
	return option(func(c *config) { c.usage = usage })
}

// Hidden marks the flag as hidden, omitting it from generated help and usage
// output while leaving it functional when specified.
func Hidden() Option {
	return option(func(c *config) { c.hidden = true })
}

// Required marks the flag as required, so parsing fails when it is omitted. It
// is shorthand for [MarkRequired] on the registered flag.
func Required() Option {
	return option(func(c *config) { c.required = true })
}

// Type overrides the reported flag type name, bypassing the default kebab-case
// name derived from the Go type.
func Type(name string) Option {
	return option(func(c *config) {
		c.typeName = func(any) string { return name }
	})
}

// DefaultFromEnv adds an environment variable that will be sourced for a default
// value for this flag.
func DefaultFromEnv(name string) Option {
	return option(func(c *config) {
		c.envs = append(c.envs, name)
	})
}

// DefaultFromFunc adds an [DefaultFunc] that will be executed for computing a
// default flag value if this was not specified.
func DefaultFromFunc(fn DefaultFunc) Option {
	return option(func(c *config) {
		c.custom = append(c.custom, fn)
	})
}

// UnmarshalWith overrides how the raw flag bytes are decoded into the flag's
// value, replacing the default of [Unmarshal]. The type parameter is inferred
// from unmarshal, so callers supply a fully typed decoder.
func UnmarshalWith[T any](unmarshal func(data []byte) (T, error)) Option {
	return option(func(c *config) {
		c.set = func(out any, data []byte) error {
			v, err := unmarshal(data)
			if err != nil {
				return err
			}
			ptr, ok := out.(*T)
			if !ok {
				return fmt.Errorf("unmarshal: %w", errDecoderType)
			}
			*ptr = v
			return nil
		}
	})
}

// Repeatable allows the flag to be specified any number of times. A repeated
// non-slice flag keeps the last value; a slice flag accumulates across
// occurrences.
func Repeatable() Option {
	return option(func(c *config) {
		c.maxSet = true
		c.maxCount = 0
	})
}

// RepeatableUpTo allows the flag to be specified up to n times, reporting
// [ErrTooManyOccurrences] on any further occurrence. It panics if n is less
// than one.
func RepeatableUpTo(n int) Option {
	if n < 1 {
		panic("flag: RepeatableUpTo requires a positive count")
	}
	return option(func(c *config) {
		c.maxSet = true
		c.maxCount = n
		c.capped = true
	})
}

// Callback registers fn to be invoked with the flag's decoded value each time
// the flag is set. fn must be a function of one of these shapes:
//
//	func()
//	func(value)
//	func() error
//	func(value) error
//
// where value is any type the decoded flag value is assignable or convertible
// to (for example func(any) on a string flag). A function returning a non-nil
// error fails parsing with that error.
//
// Callback panics if fn is not a function of a supported shape. If the decoded
// value cannot be converted to fn's parameter type, setting the flag reports an
// error wrapping [errCallbackType].
func Callback(fn any) Option {
	validateCallbackShape(fn)
	return option(func(c *config) {
		c.callbacks = append(c.callbacks, reflect.ValueOf(fn))
	})
}

// errorType is the [reflect.Type] of the error interface, used to validate
// [Callback] result signatures.
var errorType = reflect.TypeFor[error]()

// validateCallbackShape panics unless fn is a function taking at most one
// argument and returning either nothing or a single error.
func validateCallbackShape(fn any) {
	rt := reflect.TypeOf(fn)
	if rt == nil || rt.Kind() != reflect.Func {
		panic("flag: Callback requires a function")
	}
	if rt.IsVariadic() || rt.NumIn() > 1 {
		panic("flag: Callback function must take at most one argument")
	}
	switch rt.NumOut() {
	case 0:
	case 1:
		if !rt.Out(0).Implements(errorType) {
			panic("flag: Callback function must return error or nothing")
		}
	default:
		panic("flag: Callback function must return error or nothing")
	}
}

// invokeCallback calls fn with arg, converting arg to fn's parameter type when
// necessary. It returns [errCallbackType] if arg is not convertible, or any
// error fn itself returns.
func invokeCallback(fn reflect.Value, arg reflect.Value) error {
	ft := fn.Type()
	var in []reflect.Value
	if ft.NumIn() == 1 {
		p := ft.In(0)
		switch {
		case arg.Type().AssignableTo(p):
		case arg.Type().ConvertibleTo(p):
			arg = arg.Convert(p)
		default:
			return fmt.Errorf("callback: %w", errCallbackType)
		}
		in = []reflect.Value{arg}
	}
	out := fn.Call(in)
	if len(out) == 1 {
		if err, _ := out[0].Interface().(error); err != nil {
			return err
		}
	}
	return nil
}

// value is a closure-backed [pflag.Value] used by [Flag].
type value struct {
	set func(string) error
	str func() string
	typ func() string
}

func (v *value) Set(s string) error { return v.set(s) }
func (v *value) String() string     { return v.str() }
func (v *value) Type() string       { return v.typ() }

var _ pflag.Value = (*value)(nil)

// isBuiltin reports whether T is a predeclared or composite type (such as bool
// or []string) rather than a defined type such as `type Foo []string`. Defined
// types never receive the bool bare-flag default nor slice accumulation.
func isBuiltin[T any]() bool {
	return reflect.TypeFor[T]().PkgPath() == ""
}

// appendInto appends the elements of add onto the slice addressed by dst.
func appendInto[T any](dst *T, add T) {
	sv := reflect.ValueOf(dst).Elem()
	sv.Set(reflect.AppendSlice(sv, reflect.ValueOf(add)))
}

// defaultString renders out for display, dereferencing pointers and reporting an
// empty string for a nil value.
func defaultString(out any) string {
	rv := reflect.ValueOf(out)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	return fmt.Sprintf("%v", rv.Interface())
}
