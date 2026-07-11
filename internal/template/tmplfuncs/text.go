package tmplfuncs

import (
	"strings"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/richtext"
)

// Text exposes plain-text formatting helpers to templates. Its methods never
// emit colour, so callers style the result separately without disturbing width
// or alignment.
type Text struct{}

// Wrap reflows s to fit within columns of width per line. A non-positive columns
// returns s unchanged.
func (Text) Wrap(columns int, s string) string {
	return format.Resize(s, columns)
}

// Indent prefixes every line of s with n spaces.
func (Text) Indent(n int, s string) string {
	pad := strings.Repeat(" ", max(n, 0))
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}

// IndentLines joins lines with newlines and indents every line by n spaces.
func (t Text) IndentLines(n int, lines []string) string {
	return t.Indent(n, strings.Join(lines, "\n"))
}

// Pad right-pads s with spaces to width, returning s unchanged when it is
// already at least that wide. Width is measured on the visible text, so any
// styling tags in s do not count toward it.
func (Text) Pad(width int, s string) string {
	if n := width - richtext.Len(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

// Upper returns s with all letters mapped to upper case.
func (Text) Upper(s string) string {
	return strings.ToUpper(s)
}

// MaxWidth returns the visible width of the widest string in values, or zero
// when values is empty. Width is measured on the visible text, so any styling
// tags in a value do not count toward it.
func (Text) MaxWidth(values []string) int {
	width := 0
	for _, value := range values {
		width = max(width, richtext.Len(value))
	}
	return width
}
