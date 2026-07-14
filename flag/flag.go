package flag

import (
	"slices"

	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/spf13/pflag"
)

// Flag is a flag registered in a [Registry]. It exposes the properties that
// describe the flag to a user -- how it is named and documented, and the
// constraints it participates in -- while keeping the underlying flag
// representation opaque.
//
// A nil [Flag] reports the zero-value of every property.
type Flag struct {
	flag *pflag.Flag
}

// Flag is an escape-hatch to allow direct access to the underlying
// [pflag.Flag]. In general, this should _only_ be used as a transitional
// mechanism -- but should otherwise aim to avoid relying on this.
func (f *Flag) Flag() *pflag.Flag {
	if f == nil {
		return nil
	}
	return f.flag
}

// Name returns the long name of the flag, as specified on a command-line with a
// double-dash prefix.
func (f *Flag) Name() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Name
}

// Shorthand returns the single-character alias of the flag, as specified on a
// command-line with a single-dash prefix. It is empty if the flag has no alias.
func (f *Flag) Shorthand() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Shorthand
}

// Usage returns the help string displayed for the flag.
func (f *Flag) Usage() string {
	if f.Flag() == nil {
		return ""
	}
	return f.flag.Usage
}

// Type returns the name of the type the flag's value is reported as, such as
// "string" or "mips-op-code".
func (f *Flag) Type() string {
	if f.Flag() == nil || f.flag.Value == nil {
		return ""
	}
	return f.flag.Value.Type()
}

// Hidden reports whether the flag is omitted from generated help and usage
// output.
func (f *Flag) Hidden() bool {
	if f.Flag() == nil {
		return false
	}
	return f.flag.Hidden
}

// Required reports whether the flag must be specified, as marked by
// [MarkRequired] or the [Required] option.
func (f *Flag) Required() bool {
	if f.Flag() == nil {
		return false
	}
	return annotation.IsRequired(f.flag)
}

// Group returns the name of the display group the flag was added to by
// [AddToGroup]. It is empty if the flag is in no named group.
func (f *Flag) Group() string {
	if f.Flag() == nil {
		return ""
	}
	return annotation.Group(f.flag)
}

// MutuallyExclusiveWith returns the sorted names of the flags that may not be
// specified alongside this one, as marked by [MarkMutuallyExclusive]. The result
// includes this flag, and is empty if it is in no such group.
func (f *Flag) MutuallyExclusiveWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.MutuallyExclusive(f.flag)
}

// RequiredWith returns the sorted names of the flags that must be specified
// alongside this one, as marked by [MarkRequiredTogether]. The result includes
// this flag, and is empty if it is in no such group.
func (f *Flag) RequiredWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.RequiredTogether(f.flag)
}

// OneRequiredWith returns the sorted names of the flags of which at least one
// must be specified, as marked by [MarkOneRequired]. The result includes this
// flag, and is empty if it is in no such group.
func (f *Flag) OneRequiredWith() []string {
	if f.Flag() == nil {
		return nil
	}
	return annotation.OneRequired(f.flag)
}

// Equal reports whether f and other describe the same flag, comparing every
// property the flag exposes. A nil flag is equal to any other flag that
// describes nothing.
//
// This enables [Flag] values to be compared with
// [github.com/google/go-cmp/cmp.Equal].
func (f *Flag) Equal(other *Flag) bool {
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
func flagsOf(fs *pflag.FlagSet) []*Flag {
	var result []*Flag
	fs.VisitAll(func(f *pflag.Flag) {
		result = append(result, &Flag{flag: f})
	})
	return result
}
