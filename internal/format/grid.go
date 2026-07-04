package format

import (
	"strings"
	"unicode/utf8"
)

// Row is a single row of a two-column [Grid]: an aligned left marker column and
// a wrapped right description column.
type Row struct {
	// Marker is the left column, rendered verbatim. It may contain ANSI escape
	// sequences; [Grid] never inspects or parses its contents.
	Marker string

	// MarkerWidth is the visible width of Marker used for column alignment. When
	// zero, the rune count of Marker is used instead.
	MarkerWidth int

	// Description is the right column. It is plain text, wrapped to fit the width
	// remaining after the marker column.
	Description string
}

// width reports the visible width of the row's marker.
func (r Row) width() int {
	if r.MarkerWidth > 0 {
		return r.MarkerWidth
	}
	return utf8.RuneCountInString(r.Marker)
}

// Grid renders rows as a two-column layout: markers aligned to the widest
// visible marker, and descriptions wrapped at columns with continuation lines
// indented beneath the description column. indent is the leading indent applied
// to every line; gap is the number of spaces between the marker and description
// columns. minMarker is a lower bound on the marker column width, used to align
// several grids to a shared width; pass 0 to size the column to the rows alone.
//
// Grid returns the empty string when rows is empty. When the width remaining for
// descriptions is not positive (for example when columns is non-positive),
// descriptions are emitted on a single line without wrapping.
func Grid(rows []Row, columns, indent, gap, minMarker int) string {
	if len(rows) == 0 {
		return ""
	}
	markerCol := minMarker
	for _, row := range rows {
		markerCol = max(markerCol, row.width())
	}
	descCol := columns - indent - markerCol - gap
	lead := strings.Repeat(" ", atLeast(indent, 0))
	hang := strings.Repeat(" ", atLeast(indent+markerCol+gap, 0))
	sep := strings.Repeat(" ", atLeast(gap, 0))

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		pad := strings.Repeat(" ", atLeast(markerCol-row.width(), 0))
		prefix := lead + row.Marker + pad
		desc := descriptionLines(row.Description, descCol)
		if len(desc) == 0 {
			lines = append(lines, strings.TrimRight(prefix, " "))
			continue
		}
		lines = append(lines, prefix+sep+desc[0])
		for _, line := range desc[1:] {
			lines = append(lines, hang+line)
		}
	}
	return strings.Join(lines, "\n")
}

// descriptionLines wraps text to width, returning it as a single trimmed line
// when width is not positive and nil when text is blank.
func descriptionLines(text string, width int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if width <= 0 {
		return []string{strings.TrimSpace(text)}
	}
	return wrap(text, width, width)
}
