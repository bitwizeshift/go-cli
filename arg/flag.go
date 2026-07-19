package arg

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/completion"
	"github.com/spf13/pflag"
)

// FlagArg is a flag registered in a [CommandLine] with [CommandLine.Add]. It is
// produced by [Flag], and exposes the properties that describe the flag to a
// user -- how it is named and documented, and the constraints it participates in
// -- while keeping the underlying flag representation opaque.
//
// A nil [FlagArg] reports the zero-value of every property.
type FlagArg struct {
	flag *pflag.Flag
}

// Flag constructs a flag named name whose value is decoded into v. The returned
// [FlagArg] is registered on a [CommandLine] with [CommandLine.Add].
//
// By default the flag is decoded with [Unmarshal] and reports a kebab-case type
// name derived from T; both may be adjusted with [Option] values. A bool-kinded
// T is registered so that a bare --name implies true.
//
// An unnamed slice T accumulates across repeated occurrences, so --name a --name
// b is equivalent to --name a,b; any other T reports [ErrAlreadySet] if
// specified more than once. [Repeatable] lifts that limit, and [RepeatableUpTo]
// caps it, reporting [ErrTooManyOccurrences] beyond the cap; a repeated non-slice
// flag keeps the last value. [Callback] options are invoked with the decoded
// value on each occurrence.
func Flag[T any](name string, v *T, options ...FlagOption) *FlagArg {
	cfg := newFlagConfig(options...)
	slice := isBuiltin[T]() && reflect.TypeFor[T]().Kind() == reflect.Slice
	limit := 1
	if slice {
		limit = 0 // builtin slices accumulate without a limit by default
	}
	if cfg.maxSet {
		limit = cfg.maxCount
	}
	count := 0
	val := &value{
		set: func(s string) error {
			if limit > 0 && count >= limit {
				if cfg.capped {
					return fmt.Errorf("%s: %w", name, ErrTooManyOccurrences)
				}
				return fmt.Errorf("%s: %w", name, ErrAlreadySet)
			}
			var tmp T
			if err := cfg.set(&tmp, []byte(s)); err != nil {
				return err
			}
			if slice && count > 0 {
				appendInto(v, tmp)
			} else {
				*v = tmp
			}
			for _, cb := range cfg.callbacks {
				if err := invokeCallback(cb, reflect.ValueOf(tmp)); err != nil {
					return err
				}
			}
			count++
			return nil
		},
		str: func() string { return defaultString(v) },
		typ: func() string { return cfg.typeName(v) },
	}
	return newFlagArg[T](val, name, cfg)
}

// register adds the underlying flag to cl's flag set.
func (f *FlagArg) register(cl *CommandLine) {
	argdef.Flags((*argdef.CommandLine)(cl)).AddFlag(f.flag)
}

var _ Arg = (*FlagArg)(nil)

// Flag is an escape-hatch to allow direct access to the underlying
// [pflag.Flag]. In general, this should _only_ be used as a transitional
// mechanism -- but should otherwise aim to avoid relying on this.
func (f *FlagArg) Flag() *pflag.Flag {
	if f == nil {
		return nil
	}
	return f.flag
}

// Name returns the long name of the flag, as specified on a command-line with a
// double-dash prefix.
func (f *FlagArg) Name() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Name
}

// Shorthand returns the single-character alias of the flag, as specified on a
// command-line with a single-dash prefix. It is empty if the flag has no alias.
func (f *FlagArg) Shorthand() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Shorthand
}

// Usage returns the help string displayed for the flag.
func (f *FlagArg) Usage() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Usage
}

// Type returns the name of the type the flag's value is reported as, such as
// "string" or "mips-op-code".
func (f *FlagArg) Type() string {
	if f.Flag() == nil || f.flag.Value == nil {
		return ""
	}
	return f.flag.Value.Type()
}

// Hidden reports whether the flag is omitted from generated help and usage
// output.
func (f *FlagArg) Hidden() bool {
	if f.Flag() == nil {
		return false
	}
	return f.flag.Hidden
}

// Required reports whether the flag must be specified, as marked by
// [MarkRequired] or the [Required] option.
func (f *FlagArg) Required() bool {
	if f.Flag() == nil {
		return false
	}
	return annotation.IsRequired(f.flag)
}

// Group returns the name of the display group the flag was added to by
// [Group]. It is empty if the flag is in no named group.
func (f *FlagArg) Group() string {
	if f.Flag() == nil {
		return ""
	}
	return annotation.Group(f.flag)
}

// MutuallyExclusiveWith returns the sorted names of the flags that may not be
// specified alongside this one, as marked by [MarkMutuallyExclusive]. The result
// includes this flag, and is empty if it is in no such group.
func (f *FlagArg) MutuallyExclusiveWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.MutuallyExclusive(f.flag)
}

// RequiredWith returns the sorted names of the flags that must be specified
// alongside this one, as marked by [MarkRequiredTogether]. The result includes
// this flag, and is empty if it is in no such group.
func (f *FlagArg) RequiredWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.RequiredTogether(f.flag)
}

// OneRequiredWith returns the sorted names of the flags of which at least one
// must be specified, as marked by [MarkOneRequired]. The result includes this
// flag, and is empty if it is in no such group.
func (f *FlagArg) OneRequiredWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.OneRequired(f.flag)
}

// Equal reports whether f and other describe the same flag, comparing every
// property the flag exposes. A nil flag is equal to any other flag that
// describes nothing.
//
// This enables [FlagArg] values to be compared with
// [github.com/google/go-cmp/cmp.Equal].
func (f *FlagArg) Equal(other *FlagArg) bool {
	return f.Name() == other.Name() &&
		f.Shorthand() == other.Shorthand() &&
		f.Usage() == other.Usage() &&
		f.Type() == other.Type() &&
		f.Hidden() == other.Hidden() &&
		f.Required() == other.Required() &&
		f.Group() == other.Group() &&
		slices.Equal(f.MutuallyExclusiveWith(), other.MutuallyExclusiveWith()) &&
		slices.Equal(f.RequiredWith(), other.RequiredWith()) &&
		slices.Equal(f.OneRequiredWith(), other.OneRequiredWith())
}

// flagsOf wraps every flag registered in fs, in the order fs visits them.
func flagsOf(fs *pflag.FlagSet) []*FlagArg {
	var result []*FlagArg
	fs.VisitAll(func(f *pflag.Flag) {
		result = append(result, &FlagArg{flag: f})
	})
	return result
}

// newFlagArg builds the underlying [pflag.Flag] for val and applies every
// configured annotation, without inserting it into any flag set. Insertion is
// deferred to [FlagArg.register]. It applies the bool bare-flag default for an
// unnamed bool T.
func newFlagArg[T any](val *value, name string, cfg *flagConfig) *FlagArg {
	f := &pflag.Flag{
		Name:      name,
		Shorthand: cfg.shorthand,
		Usage:     cfg.usage,
		Value:     val,
		DefValue:  val.String(),
	}
	f.Hidden = cfg.hidden
	if cfg.required {
		annotation.MarkRequired(f)
	}
	if isBuiltin[T]() && reflect.TypeFor[T]().Kind() == reflect.Bool {
		f.NoOptDefVal = "true"
	}
	for _, env := range cfg.envs {
		annotation.AddEnvFallback(f, env)
	}
	for _, fn := range cfg.custom {
		annotation.AddFuncFallback(f, fn)
	}
	if cfg.completer != nil {
		completion.AddFlag(f, cfg.completer)
	}
	return &FlagArg{flag: f}
}

// MarkRequired marks that all of the specified flags must be required when
// parsing command lines.
func MarkRequired(flags ...*FlagArg) {
	annotation.MarkRequired(pflags(flags)...)
}

// MarkRequiredTogether marks that all flags must be specified together when any
// one flag is specified. Note that this does not mean that all flags are always
// required; it's all or none. If all flags are always required, then
// [MarkRequired] should be used.
func MarkRequiredTogether(flags ...*FlagArg) {
	annotation.MarkRequiredTogether(pflags(flags)...)
}

// MarkMutuallyExclusive marks that all flags must be mutually exclusive with
// each other, and will generate an error when parsing flags that have both set.
func MarkMutuallyExclusive(flags ...*FlagArg) {
	annotation.MarkMutuallyExclusive(pflags(flags)...)
}

// MarkOneRequired marks that at least one of the specified flags is required
// when parsing command lines.
func MarkOneRequired(flags ...*FlagArg) {
	annotation.MarkOneRequired(pflags(flags)...)
}

// pflags unwraps flags into the representation the annotations are recorded on.
func pflags(flags []*FlagArg) []*pflag.Flag {
	result := make([]*pflag.Flag, 0, len(flags))
	for _, f := range flags {
		result = append(result, f.Flag())
	}
	return result
}
