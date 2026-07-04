package help

import (
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
)

// Layout constants shared by the help sections.
const (
	sectionIndent = 2
	gridGap       = 3
)

// span accumulates a rendered marker alongside its visible width, measured from
// the plain tokens only so that colour never perturbs column alignment.
type span struct {
	b     strings.Builder
	width int
}

// add appends text styled by paint, adding only the plain rune count of text to
// the visible width.
func (s *span) add(text string, paint func(string) string) {
	s.b.WriteString(paint(text))
	s.width += utf8.RuneCountInString(text)
}

// row returns the accumulated marker as a [format.Row] with the given
// description.
func (s *span) row(description string) format.Row {
	return format.Row{Marker: s.b.String(), MarkerWidth: s.width, Description: description}
}

// maxCommandColumn caps the shared command name column so that a very long name
// only widens its own group rather than every group.
const maxCommandColumn = 16

// commandColumnWidth returns the shared width for command name columns across
// groups: the longest command name in any group, capped at [maxCommandColumn].
func commandColumnWidth(groups []CommandGroup) int {
	longest := 0
	for _, group := range groups {
		for _, c := range group.Commands {
			longest = max(longest, utf8.RuneCountInString(c.Name))
		}
	}
	return min(longest, maxCommandColumn)
}

// commandGrid renders subcommands as an aligned grid, colouring each name with p.
// The name column is at least width wide so that groups share an alignment.
func commandGrid(p palette.Palette, commands []Command, columns, width int) string {
	rows := make([]format.Row, 0, len(commands))
	for _, c := range commands {
		var s span
		s.add(c.Name, p.Title)
		rows = append(rows, s.row(c.Summary))
	}
	return format.Grid(rows, columns, sectionIndent, gridGap, width)
}

// flagGrid renders flags as an aligned grid, colouring flag names and their type
// arguments distinctly with p.
func flagGrid(p palette.Palette, flags []FlagInfo, columns int) string {
	rows := make([]format.Row, 0, len(flags))
	for _, f := range flags {
		var s span
		s.add(flagMarker(f), p.Label)
		if f.Type != "" {
			s.add(" "+f.Type, p.Value)
		}
		rows = append(rows, s.row(f.Usage))
	}
	return format.Grid(rows, columns, sectionIndent, gridGap, 0)
}

// flagMarker renders the name column of a flag, aligning long-only flags beneath
// the position occupied by a shorthand.
func flagMarker(f FlagInfo) string {
	if f.Shorthand != "" {
		return "-" + f.Shorthand + ", --" + f.Name
	}
	return "    --" + f.Name
}

// hintSuffix completes the trailing help advice after the emphasised command.
const hintSuffix = " for more information about a command."

// hintLine renders the trailing "--help" advice for a command at path, wrapped
// to columns. The command phrase is emphasised after wrapping so that its colour
// codes never affect the wrap width; if wrapping splits the phrase the emphasis
// is a no-op rather than corrupting the line.
func hintLine(p palette.Palette, path string, columns int) string {
	phrase := path + " [command] --help"
	wrapped := format.Resize("Use "+phrase+hintSuffix, columns)
	return strings.Replace(wrapped, phrase, p.Emphasis(phrase), 1)
}

// funcs builds the template function map for rendering view at the given width
// using palette p. It extends the shared [tmplfuncs.NewFunc] set with the help
// grid and hint layout functions.
func funcs(columns int, p palette.Palette, view View) template.FuncMap {
	commandWidth := commandColumnWidth(view.CommandGroups)
	f := tmplfuncs.NewFunc(p)
	f["commandGrid"] = func(commands []Command) string {
		return commandGrid(p, commands, columns, commandWidth)
	}
	f["flagGrid"] = func(flags []FlagInfo) string { return flagGrid(p, flags, columns) }
	f["hint"] = func(path string) string { return hintLine(p, path, columns) }
	return f
}
