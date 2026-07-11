package style

import (
	"fmt"

	"github.com/bitwizeshift/go-cli/richtext/internal/sgr"
)

// Style is a complete set of terminal styling: foreground colour, background
// colour, and text attributes. Its zero value applies no styling and renders as
// an empty string.
type Style struct {
	Foreground Colour
	Background Colour
	Attributes Attribute
}

// Merge overlays other onto s and returns the result. A set colour in other
// replaces the corresponding colour in s; attributes combine with bitwise OR.
func (s Style) Merge(other Style) Style {
	if other.Foreground.isSet() {
		s.Foreground = other.Foreground
	}
	if other.Background.isSet() {
		s.Background = other.Background
	}
	s.Attributes |= other.Attributes
	return s
}

// String returns the SGR escape sequence for s, without a leading reset. The
// zero Style returns the empty string.
func (s Style) String() string {
	params := s.Attributes.params()
	params = append(params, s.Foreground.params(false)...)
	params = append(params, s.Background.params(true)...)
	return sgr.Sequence(params...)
}

var _ fmt.Stringer = Style{}
