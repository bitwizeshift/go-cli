package ansi

import (
	"fmt"
)

// SGRParam represents an ANSI SGR attribute.
type SGRParam uint8

// Format returns the ANSI SGR escape sequence for the attribute.
func (p SGRParam) Format(w fmt.State, verb rune) {
	switch verb {
	case 'v':
		fmt.Fprintf(w, "%s%d%s", SGRPrefix, uint8(p), SGRSuffix)
	case 'd':
		fmt.Fprintf(w, "%d", uint8(p))
	default:
		fmt.Fprintf(w, "!%c(SGRParam=%d)", verb, p)
	}
}

// String returns the ANSI SGR escape sequence for the attribute.
func (p SGRParam) String() string {
	return fmt.Sprintf("%v", p)
}

var (
	_ fmt.Formatter = (*SGRParam)(nil)
	_ fmt.Stringer  = (*SGRParam)(nil)
)

const (
	// Reset resets all attributes.
	Reset SGRParam = 0

	// Bold makes text bold.
	Bold SGRParam = 1

	// Dim makes text dim.
	Dim SGRParam = 2

	// Italic makes text italic.
	Italic SGRParam = 3

	// Underline makes text underlined.
	Underline SGRParam = 4

	//--------------------------------------------------------------------------

	// FGBlack sets the foreground color to black.
	FGBlack SGRParam = 30

	// FGRed sets the foreground color to red.
	FGRed SGRParam = 31

	// FGGreen sets the foreground color to green.
	FGGreen SGRParam = 32

	// FGYellow sets the foreground color to yellow.
	FGYellow SGRParam = 33

	// FGBlue sets the foreground color to blue.
	FGBlue SGRParam = 34

	// FGMagenta sets the foreground color to magenta.
	FGMagenta SGRParam = 35

	// FGCyan sets the foreground color to cyan.
	FGCyan SGRParam = 36

	// FGWhite sets the foreground color to white.
	FGWhite SGRParam = 37

	// FGDefault resets the foreground color to the default.
	FGDefault SGRParam = 39

	//--------------------------------------------------------------------------

	// BGBlack sets the background color to black.
	BGBlack SGRParam = 40

	// BGRed sets the background color to red.
	BGRed SGRParam = 41

	// BGGreen sets the background color to green.
	BGGreen SGRParam = 42

	// BGYellow sets the background color to yellow.
	BGYellow SGRParam = 43

	// BGBlue sets the background color to blue.
	BGBlue SGRParam = 44

	// BGMagenta sets the background color to magenta.
	BGMagenta SGRParam = 45

	// BGCyan sets the background color to cyan.
	BGCyan SGRParam = 46

	// BGWhite sets the background color to white.
	BGWhite SGRParam = 47

	// BGDefault resets the background color to the default.
	BGDefault SGRParam = 49

	//--------------------------------------------------------------------------

	// FGBrightBlack sets the foreground color to bright black.
	FGBrightBlack SGRParam = 90

	// FGBrightRed sets the foreground color to bright red.
	FGBrightRed SGRParam = 91

	// FGBrightGreen sets the foreground color to bright green.
	FGBrightGreen SGRParam = 92

	// FGBrightYellow sets the foreground color to bright yellow.
	FGBrightYellow SGRParam = 93

	// FGBrightBlue sets the foreground color to bright blue.
	FGBrightBlue SGRParam = 94

	// FGBrightMagenta sets the foreground color to bright magenta.
	FGBrightMagenta SGRParam = 95

	// FGBrightCyan sets the foreground color to bright cyan.
	FGBrightCyan SGRParam = 96

	// FGBrightWhite sets the foreground color to bright white.
	FGBrightWhite SGRParam = 97

	//--------------------------------------------------------------------------

	// BGBrightBlack sets the background color to bright black.
	BGBrightBlack SGRParam = 100

	// BGBrightRed sets the background color to bright red.
	BGBrightRed SGRParam = 101

	// BGBrightGreen sets the background color to bright green.
	BGBrightGreen SGRParam = 102

	// BGBrightYellow sets the background color to bright yellow.
	BGBrightYellow SGRParam = 103

	// BGBrightBlue sets the background color to bright blue.
	BGBrightBlue SGRParam = 104

	// BGBrightMagenta sets the background color to bright magenta.
	BGBrightMagenta SGRParam = 105

	// BGBrightCyan sets the background color to bright cyan.
	BGBrightCyan SGRParam = 106

	// BGBrightWhite sets the background color to bright white.
	BGBrightWhite SGRParam = 107
)

// Format is a convenience function to create an [SGRControlSequence] from a
// list of [SGRParam] attributes.
func Format(params ...SGRParam) SGRControlSequence {
	return SGRControlSequence(params)
}
