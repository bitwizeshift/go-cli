package sgr

import (
	"strconv"
	"strings"
)

// Reset is the SGR sequence that clears every previously-set attribute.
const Reset = "\x1b[0m"

// backgroundOffset is the fixed distance between a foreground colour parameter
// and its background counterpart in the SGR parameter space (e.g. 31 -> 41,
// 91 -> 101).
const backgroundOffset = 10

// TrueColour returns the SGR parameters selecting a 24-bit colour. When bg is
// true the parameters target the background layer, otherwise the foreground.
func TrueColour(bg bool, r, g, b uint8) []int {
	lead := 38
	if bg {
		lead = 48
	}
	return []int{lead, 2, int(r), int(g), int(b)}
}

// Background shifts a foreground colour parameter to its background equivalent.
func Background(foreground int) int {
	return foreground + backgroundOffset
}

// Sequence renders params as a single SGR escape sequence. It returns the empty
// string when no parameters are given.
func Sequence(params ...int) string {
	if len(params) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\x1b[")
	for i, p := range params {
		if i > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(p))
	}
	b.WriteByte('m')
	return b.String()
}
