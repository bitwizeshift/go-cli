package argreg

import (
	"unsafe"

	"github.com/spf13/pflag"
)

// CommandLine is the opaque command-line destination threaded through
// registration.
//
// This type forms the base type for the proper [arg.CommandLine]
type CommandLine struct {
	flags       *pflag.FlagSet
	visited     map[unsafe.Pointer]struct{}
	positionals []*Positional
	unmatched   *Unmatched
}

// Positional is a registered positional-argument binding.
type Positional struct {
	Index int
	Name  string
	Type  string
	Usage string
	Set   func(value string) error
}

// Unmatched is a registered binding for every argument not claimed by a
// [Positional].
type Unmatched struct {
	Set func(values []string) error
}

// New returns a newly constructed [CommandLine]. This is to enable creating
// registries for testing purposes.
func New() *CommandLine {
	flags := pflag.NewFlagSet("registry", pflag.ContinueOnError)
	return FromFlagSet(flags)
}

// FromFlagSet constructs a [CommandLine] from a [pflag.FlagSet]. This is used in
// the real CLI construction.
func FromFlagSet(flags *pflag.FlagSet) *CommandLine {
	return &CommandLine{flags: flags, visited: map[unsafe.Pointer]struct{}{}}
}

// The below functions exist so that other exported packages are able to access
// unexported fields in their implementation.

// Flags returns the flags for this registry.
func Flags(reg *CommandLine) *pflag.FlagSet {
	return reg.flags
}

// Visited returns the map for visited entries for this registry.
func Visited(reg *CommandLine) map[unsafe.Pointer]struct{} {
	return reg.visited
}

// AddPositional records p as a positional-argument binding on reg.
func AddPositional(reg *CommandLine, p *Positional) {
	reg.positionals = append(reg.positionals, p)
}

// Positionals returns the positional-argument bindings registered on reg, in
// registration order.
func Positionals(reg *CommandLine) []*Positional {
	return reg.positionals
}

// SetUnmatched records u as the unmatched-argument binding on reg, replacing any
// previously registered binding.
func SetUnmatched(reg *CommandLine, u *Unmatched) {
	reg.unmatched = u
}

// GetUnmatched returns the unmatched-argument binding registered on reg, or nil
// if none was registered.
func GetUnmatched(reg *CommandLine) *Unmatched {
	return reg.unmatched
}

// Bind assigns args to the positional and unmatched bindings registered on reg.
// A [Positional] whose index falls outside args is skipped, leaving its
// destination unchanged. Every argument not claimed by a [Positional] is passed,
// in command-line order, to an [Unmatched] binding.
//
// It returns the first error reported by a binding.
func Bind(reg *CommandLine, args []string) error {
	claimed := make(map[int]struct{})
	for _, p := range reg.positionals {
		if p.Index < 0 || p.Index >= len(args) {
			continue
		}
		claimed[p.Index] = struct{}{}
		if err := p.Set(args[p.Index]); err != nil {
			return err
		}
	}
	if reg.unmatched == nil {
		return nil
	}
	rest := make([]string, 0, len(args))
	for i, arg := range args {
		if _, ok := claimed[i]; ok {
			continue
		}
		rest = append(rest, arg)
	}
	return reg.unmatched.Set(rest)
}
