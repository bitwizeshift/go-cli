package arg

import (
	"github.com/bitwizeshift/go-cli/internal/argdef"
)

// FlagGroup represents a grouping of flags denoted by the name of a given group
type FlagGroup struct {
	// Name is the name of the flag group.
	Name string

	// Flags is a list of all the flags that are part of this group.
	Flags []*FlagArg
}

// Hidden returns true if all flags in the group are marked as Hidden.
func (g *FlagGroup) Hidden() bool {
	allHidden := true
	for _, f := range g.Flags {
		allHidden = allHidden && f.Hidden()
	}
	return allHidden
}

// Group assigns each of the given flags to the named display group. Flags
// that share a group name are reported together by [CommandLine.Groups], and
// rendered under that heading in help output. A flag that is not added to any
// group is reported under the default "General Flags" heading.
func Group(name string, flags ...*FlagArg) {
	argdef.AddToGroup(name, pflags(flags)...)
}
