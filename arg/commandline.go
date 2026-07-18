package arg

import (
	"slices"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/argreg"
	"github.com/spf13/pflag"
)

// CommandLine is the opaque flag destination threaded through registration. It wraps
// a [pflag.CommandLine] and records which [Registrar] instances have already been
// registered so that a shared instance registers its flags only once.
// [CommandLine.Add] is its only mutator.
type CommandLine argreg.CommandLine

// FlagSet is an escape-hatch to allow direct access to the underlying
// [flag.FlagSet]. In general, this should _only_ be used as a transitional
// mechanism -- but should otherwise aim to avoid relying on this.
func (r *CommandLine) FlagSet() *pflag.FlagSet {
	return argreg.Flags((*argreg.CommandLine)(r))
}

// Flags returns every [FlagArg] registered in this registry, in lexical order.
func (r *CommandLine) Flags() []*FlagArg {
	return flagsOf(r.FlagSet())
}

// Add registers each argument on the command line, in the order given.
func (r *CommandLine) Add(args ...Arg) {
	for _, a := range args {
		a.register(r)
	}
}

// AddFlagSet registers every flag in fs on the command line.
func (r *CommandLine) AddFlagSet(fs *pflag.FlagSet) {
	r.FlagSet().AddFlagSet(fs)
}

const generalFlagsGroup = "General Flags"

// Groups returns a slice containing all flag [FlagGroup]s in the registry, sorted
// by group name and subsorted by flag name. Flags that are not part of a named
// group get added to the group "General Flags" -- which is always sorted last.
func (r *CommandLine) Groups() []*FlagGroup {
	var result []*FlagGroup
	dedup := map[string]*FlagGroup{}
	for _, f := range r.Flags() {
		name := f.Group()
		if name == "" {
			name = generalFlagsGroup
		}
		g, ok := dedup[name]
		if !ok {
			g = &FlagGroup{
				Name: name,
			}
			dedup[name] = g
			result = append(result, g)
		}
		g.Flags = append(g.Flags, f)
	}
	for _, g := range result {
		slices.SortFunc(g.Flags, func(lhs, rhs *FlagArg) int {
			return strings.Compare(lhs.Name(), rhs.Name())
		})
	}
	slices.SortFunc(result, func(lhs, rhs *FlagGroup) int {
		if lhs.Name == generalFlagsGroup {
			return 1
		}
		if rhs.Name == generalFlagsGroup {
			return -1
		}
		return strings.Compare(lhs.Name, rhs.Name)
	})
	return result
}
