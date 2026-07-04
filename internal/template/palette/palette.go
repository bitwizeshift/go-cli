package palette

import "github.com/bitwizeshift/go-cli/internal/ansi"

// Palette styles the individual roles of rendered output. Each method colours
// one role and returns the styled string; a plain implementation returns its
// input unchanged.
type Palette interface {
	// Title styles a command title or name.
	Title(s string) string

	// Heading styles a section heading.
	Heading(s string) string

	// Label styles a key column, such as a flag name or a build-detail label.
	Label(s string) string

	// Value styles a value column, such as a flag's type argument or a
	// build-detail value.
	Value(s string) string

	// Emphasis styles an emphasised inline phrase.
	Emphasis(s string) string

	// Error styles an error label.
	Error(s string) string

	// Quote styles quoted or verbatim text, such as a stack-trace line.
	Quote(s string) string

	// Gutter styles the prefix that indents a quoted block.
	Gutter(s string) string

	// URL styles a hyperlink.
	URL(s string) string
}

// NoColour is a [Palette] that returns every string unchanged, producing plain
// uncoloured output.
type NoColour struct{}

func (NoColour) Title(s string) string    { return s }
func (NoColour) Heading(s string) string  { return s }
func (NoColour) Label(s string) string    { return s }
func (NoColour) Value(s string) string    { return s }
func (NoColour) Emphasis(s string) string { return s }
func (NoColour) Error(s string) string    { return s }
func (NoColour) Quote(s string) string    { return s }
func (NoColour) Gutter(s string) string   { return s }
func (NoColour) URL(s string) string      { return s }

var _ Palette = (*NoColour)(nil)

// Colour is a [Palette] that wraps each role in a configurable ANSI colour.
type Colour struct {
	TitleColour    ansi.Colour
	HeadingColour  ansi.Colour
	LabelColour    ansi.Colour
	ValueColour    ansi.Colour
	EmphasisColour ansi.Colour
	ErrorColour    ansi.Colour
	QuoteColour    ansi.Colour
	GutterColour   ansi.Colour
	URLColour      ansi.Colour
}

func (c Colour) Title(s string) string    { return paint(c.TitleColour, s) }
func (c Colour) Heading(s string) string  { return paint(c.HeadingColour, s) }
func (c Colour) Label(s string) string    { return paint(c.LabelColour, s) }
func (c Colour) Value(s string) string    { return paint(c.ValueColour, s) }
func (c Colour) Emphasis(s string) string { return paint(c.EmphasisColour, s) }
func (c Colour) Error(s string) string    { return paint(c.ErrorColour, s) }
func (c Colour) Quote(s string) string    { return paint(c.QuoteColour, s) }
func (c Colour) Gutter(s string) string   { return paint(c.GutterColour, s) }
func (c Colour) URL(s string) string      { return paint(c.URLColour, s) }

var _ Palette = (*Colour)(nil)

// DefaultColour is the standard colour scheme: green titles, yellow headings,
// cyan labels, bold-white values, bold emphasis, a red error label, gray quoted
// text, white gutters, and an underlined bright-white URL.
var DefaultColour = Colour{
	TitleColour:    ansi.Green,
	HeadingColour:  ansi.Yellow,
	LabelColour:    ansi.Cyan,
	ValueColour:    ansi.White,
	EmphasisColour: ansi.Bold,
	ErrorColour:    ansi.Red,
	QuoteColour:    ansi.BrightBlack,
	GutterColour:   ansi.White,
	URLColour:      ansi.Colour(string(ansi.Underline) + string(ansi.BrightWhite)),
}

// paint wraps s in colour, resetting style afterwards.
func paint(colour ansi.Colour, s string) string {
	return string(colour) + s + string(ansi.Reset)
}
