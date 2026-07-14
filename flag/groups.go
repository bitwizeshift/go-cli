package flag

import (
	"slices"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// Group represents a grouping of flags denoted by the name of a given group
type Group struct {
	// Name is the name of the flag group.
	Name string

	// Flags is a list of all the flags that are part of this group.
	Flags []*Flag
}

// Hidden returns true if all flags in the group are marked as Hidden.
func (g *Group) Hidden() bool {
	allHidden := true
	for _, f := range g.Flags {
		allHidden = allHidden && f.Hidden()
	}
	return allHidden
}

// AddToGroup assigns each of the given flags to the named display group. Flags
// that share a group name are reported together by [Groups], and rendered under
// that heading in help output. A flag that is not added to any group is reported
// under the default "General Flags" heading.
func AddToGroup(name string, flags ...*Flag) {
	annotation.AddToGroup(name, pflags(flags)...)
}

const generalFlagsGroup = "General Flags"

// Groups returns a slice containing all flag [Group]s in the registry, sorted
// by group name and subsorted by flag name. Flags that are not part of a named
// group get added to the group "General Flags" -- which is always sorted last.
func Groups(registry *Registry) []*Group {
	var result []*Group
	dedup := map[string]*Group{}
	for _, f := range registry.Flags() {
		name := f.Group()
		if name == "" {
			name = generalFlagsGroup
		}
		g, ok := dedup[name]
		if !ok {
			g = &Group{
				Name: name,
			}
			dedup[name] = g
			result = append(result, g)
		}
		g.Flags = append(g.Flags, f)
	}
	for _, g := range result {
		slices.SortFunc(g.Flags, func(lhs, rhs *Flag) int {
			return strings.Compare(lhs.Name(), rhs.Name())
		})
	}
	slices.SortFunc(result, func(lhs, rhs *Group) int {
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
