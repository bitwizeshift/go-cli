package style

import (
	"encoding"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bitwizeshift/go-cli/richtext/internal/sgr"
)

// ErrUnknownColour reports that a colour name could not be resolved.
var ErrUnknownColour = errors.New("unknown colour")

// colourKind distinguishes the way a [Colour] selects its value.
type colourKind uint8

const (
	colourUnset colourKind = iota
	colourNamed
	colourTrue
)

// Colour is a foreground or background colour. Its zero value is unset, meaning
// it contributes no styling and is overridable by [Style.Merge].
//
// A Colour is either one of the sixteen named ANSI colours (see the package
// preset variables) or a 24-bit value produced by [RGB].
type Colour struct {
	kind    colourKind
	code    int // foreground SGR parameter when kind == colourNamed
	r, g, b uint8
}

// The sixteen named ANSI colours.
var (
	Black   = named(30)
	Red     = named(31)
	Green   = named(32)
	Yellow  = named(33)
	Blue    = named(34)
	Magenta = named(35)
	Cyan    = named(36)
	White   = named(37)

	BrightBlack   = named(90)
	BrightRed     = named(91)
	BrightGreen   = named(92)
	BrightYellow  = named(93)
	BrightBlue    = named(94)
	BrightMagenta = named(95)
	BrightCyan    = named(96)
	BrightWhite   = named(97)
)

func named(code int) Colour {
	return Colour{kind: colourNamed, code: code}
}

// RGB returns a 24-bit true-colour.
func RGB(r, g, b uint8) Colour {
	return Colour{kind: colourTrue, r: r, g: g, b: b}
}

var colourNames = map[string]Colour{
	"black":         Black,
	"red":           Red,
	"green":         Green,
	"yellow":        Yellow,
	"blue":          Blue,
	"magenta":       Magenta,
	"cyan":          Cyan,
	"white":         White,
	"brightblack":   BrightBlack,
	"brightred":     BrightRed,
	"brightgreen":   BrightGreen,
	"brightyellow":  BrightYellow,
	"brightblue":    BrightBlue,
	"brightmagenta": BrightMagenta,
	"brightcyan":    BrightCyan,
	"brightwhite":   BrightWhite,
}

// ColourByName resolves one of the sixteen named ANSI colours. The second
// result reports whether the name was recognised.
func ColourByName(name string) (Colour, bool) {
	c, ok := colourNames[name]
	return c, ok
}

// UnmarshalText resolves a named colour ("red", "brightred") or a true-colour
// in "rgb(r,g,b)" form with each component in the range 0-255. It returns
// [ErrUnknownColour] for any other input.
func (c *Colour) UnmarshalText(b []byte) error {
	if named, ok := ColourByName(string(b)); ok {
		*c = named
		return nil
	}
	if rgb, ok := parseRGB(string(b)); ok {
		*c = rgb
		return nil
	}
	return fmt.Errorf("colour: %w %q", ErrUnknownColour, b)
}

func parseRGB(s string) (Colour, bool) {
	inner, ok := strings.CutPrefix(s, "rgb(")
	if !ok {
		return Colour{}, false
	}
	inner, ok = strings.CutSuffix(inner, ")")
	if !ok {
		return Colour{}, false
	}
	parts := strings.Split(inner, ",")
	if len(parts) != 3 {
		return Colour{}, false
	}
	var v [3]uint8
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 0 || n > 255 {
			return Colour{}, false
		}
		v[i] = uint8(n)
	}
	return RGB(v[0], v[1], v[2]), true
}

// isSet reports whether the colour contributes styling.
func (c Colour) isSet() bool {
	return c.kind != colourUnset
}

// params returns the SGR parameters selecting this colour on the requested
// layer. It returns nil when the colour is unset.
func (c Colour) params(background bool) []int {
	switch c.kind {
	case colourNamed:
		code := c.code
		if background {
			code = sgr.Background(code)
		}
		return []int{code}
	case colourTrue:
		return sgr.TrueColour(background, c.r, c.g, c.b)
	default:
		return nil
	}
}

var _ encoding.TextUnmarshaler = (*Colour)(nil)
