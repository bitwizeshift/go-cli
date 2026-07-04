package flag

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/strcase"
	"github.com/spf13/pflag"
)

// ErrAlreadySet indicates a non-accumulating flag was specified more than once.
// Only unnamed slice flags (such as []string) accumulate across occurrences; any
// other flag reports this error on its second occurrence.
var ErrAlreadySet = errors.New("flag already specified")

// errDecoderType indicates that a decoder supplied via [UnmarshalWith] produces
// a value whose type does not match the flag's destination.
var errDecoderType = errors.New("decoder type does not match flag value")

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

// Option configures an optional property of a flag registered by [Add] or
// [AddCallback].
type Option interface {
	apply(*config)
}

type option func(*config)

func (o option) apply(c *config) { o(c) }

// config holds the resolved options for a single flag registration.
type config struct {
	shorthand string
	usage     string
	typeName  func(any) string
	set       func(any, []byte) error
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

// Type overrides the reported flag type name, bypassing the default kebab-case
// name derived from the Go type.
func Type(name string) Option {
	return option(func(c *config) {
		c.typeName = func(any) string { return name }
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

// value is a closure-backed [pflag.Value] shared by [Add] and [AddCallback].
type value struct {
	set func(string) error
	str func() string
	typ func() string
}

func (v *value) Set(s string) error { return v.set(s) }
func (v *value) String() string     { return v.str() }
func (v *value) Type() string       { return v.typ() }

var _ pflag.Value = (*value)(nil)

// Add registers a flag named name whose value is decoded into v, returning the
// created [pflag.Flag].
//
// By default the flag is decoded with [Unmarshal] and reports a kebab-case type
// name derived from T; both may be adjusted with [Option] values. A bool-kinded
// T is registered so that a bare --name implies true. An unnamed slice T
// accumulates across repeated occurrences, so --name a --name b is equivalent
// to --name a,b; any other T reports [ErrAlreadySet] if specified more than
// once.
func Add[T any](fs *pflag.FlagSet, name string, v *T, options ...Option) *pflag.Flag {
	cfg := newConfig(options...)
	slice := isBuiltin[T]() && reflect.TypeFor[T]().Kind() == reflect.Slice
	changed := false
	val := &value{
		set: func(s string) error {
			if changed && !slice {
				return fmt.Errorf("%s: %w", name, ErrAlreadySet)
			}
			var tmp T
			if err := cfg.set(&tmp, []byte(s)); err != nil {
				return err
			}
			if slice && changed {
				appendInto(v, tmp)
			} else {
				*v = tmp
			}
			changed = true
			return nil
		},
		str: func() string { return defaultString(v) },
		typ: func() string { return cfg.typeName(v) },
	}
	return registerFlag[T](fs, val, name, cfg)
}

// AddCallback registers a flag named name that, when set, decodes its value into
// a T and invokes cb with it, returning the created [pflag.Flag].
//
// AddCallback is the functional form of [Add] and honors the same [Option]
// values. cb is invoked once per occurrence of the flag; a bool-kinded T allows
// a bare --name to invoke cb with true.
func AddCallback[T any](fs *pflag.FlagSet, name string, cb func(T) error, options ...Option) *pflag.Flag {
	cfg := newConfig(options...)
	val := &value{
		set: func(s string) error {
			var out T
			if err := cfg.set(&out, []byte(s)); err != nil {
				return err
			}
			return cb(out)
		},
		str: func() string { return "" },
		typ: func() string { return cfg.typeName((*T)(nil)) },
	}
	return registerFlag[T](fs, val, name, cfg)
}

// register adds val to fs under name, applying the bool bare-flag default for an
// unnamed bool T.
func registerFlag[T any](fs *pflag.FlagSet, val *value, name string, cfg *config) *pflag.Flag {
	f := fs.VarPF(val, name, cfg.shorthand, cfg.usage)
	if isBuiltin[T]() && reflect.TypeFor[T]().Kind() == reflect.Bool {
		f.NoOptDefVal = "true"
	}
	return f
}

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
