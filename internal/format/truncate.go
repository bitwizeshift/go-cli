package format

const ellipsis = "…"

// Truncate shortens s to at most width visible columns, replacing the trailing
// runes with a single-rune ellipsis when it must cut. It returns s unchanged
// when it already fits or when width <= 0. It counts runes, not bytes, and does
// not interpret ANSI escape sequences.
func Truncate(s string, width int) string {
	if width <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return ellipsis
	}
	return string(runes[:width-1]) + ellipsis
}
