package progress

import "strings"

// Renderer renders its current visual state to a single- or multi-line string.
// Implementations are pure: identical field state always yields the same
// string.
type Renderer interface {
	Render() string
}

// Group renders several Renderers stacked vertically, one per line. Group is
// itself a [Renderer], so groups compose into larger blocks such as a table of
// bars.
type Group []Renderer

// Render joins the renders of each member with newlines, in order. It returns
// the empty string when the group is empty.
func (g Group) Render() string {
	lines := make([]string, len(g))
	for i, r := range g {
		lines[i] = r.Render()
	}
	return strings.Join(lines, "\n")
}

var _ Renderer = (*Group)(nil)
