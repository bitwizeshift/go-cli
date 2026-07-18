package help

import (
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/template/tag"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/richtext"
)

// Layout constants shared by the help sections.
const (
	sectionIndent = 2
	gridGap       = 3
)

// span accumulates a styled marker, measuring its visible width from the markup
// so that styling tags never perturb column alignment.
type span struct {
	b strings.Builder
}

// add appends text styled with the theme role to the marker.
func (s *span) add(text, role string) {
	s.b.WriteString(tag.Themed(role, text))
}

// row returns the accumulated marker as a [format.Row] with the given
// description, sizing the marker column from its visible width.
func (s *span) row(description string) format.Row {
	marker := s.b.String()
	return format.Row{Marker: marker, MarkerWidth: richtext.Len(marker), Description: description}
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

// commandGrid renders subcommands as an aligned grid, styling each name. The
// name column is at least width wide so that groups share an alignment.
func commandGrid(commands []Command, columns, width int) string {
	rows := make([]format.Row, 0, len(commands))
	for _, c := range commands {
		var s span
		s.add(c.Name, "title")
		rows = append(rows, s.row(c.Summary))
	}
	return format.Grid(rows, columns, sectionIndent, gridGap, width)
}

// flagGrid renders flags as an aligned grid, styling flag names and their type
// arguments distinctly.
func flagGrid(flags []FlagInfo, columns int) string {
	rows := make([]format.Row, 0, len(flags))
	for _, f := range flags {
		var s span
		s.add(flagMarker(f), "label")
		if f.Type != "" {
			s.add(" "+f.Type, "value")
		}
		rows = append(rows, s.row(f.Usage))
	}
	return format.Grid(rows, columns, sectionIndent, gridGap, 0)
}

// positionalGrid renders positional arguments as an aligned grid, styling names
// and their type arguments distinctly, matching the flag grid's layout.
func positionalGrid(positionals []PositionalInfo, columns int) string {
	rows := make([]format.Row, 0, len(positionals))
	for _, p := range positionals {
		var s span
		s.add(p.Name, "label")
		if p.Type != "" {
			s.add(" "+p.Type, "value")
		}
		rows = append(rows, s.row(p.Usage))
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
// to columns. The command phrase is emphasised after wrapping so that its
// styling tags never affect the wrap width; if wrapping splits the phrase the
// emphasis is a no-op rather than corrupting the line.
func hintLine(path string, columns int) string {
	phrase := path + " [command] --help"
	wrapped := format.Resize("Use "+phrase+hintSuffix, columns)
	return strings.Replace(wrapped, phrase, tag.Themed("emphasis", phrase), 1)
}

// funcs builds the template function map for rendering view at the given width.
// It extends the shared [tmplfuncs.NewFunc] set with the help grid and hint
// layout functions.
func funcs(columns int, view View) template.FuncMap {
	commandWidth := commandColumnWidth(view.CommandGroups)
	f := tmplfuncs.NewFunc()
	f["commandGrid"] = func(commands []Command) string {
		return commandGrid(commands, columns, commandWidth)
	}
	f["flagGrid"] = func(flags []FlagInfo) string { return flagGrid(flags, columns) }
	f["positionalGrid"] = func(positionals []PositionalInfo) string {
		return positionalGrid(positionals, columns)
	}
	f["hint"] = func(path string) string { return hintLine(path, columns) }
	return f
}
