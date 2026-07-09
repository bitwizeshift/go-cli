package progress

import (
	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/term/ansi"
)

// Spinner is a renderable spinner animation. Its zero value spins the ASCII
// [LineFrames]; set fields to customise frames, colour, and postfix label. State
// advances only via [Spinner.Tick], so [Spinner.Render] is pure.
type Spinner struct {
	// Frames are the animation cells, cycled by Tick. An empty Frames defaults
	// to [LineFrames].
	Frames []string

	// Frame is the index of the currently displayed cell.
	Frame int

	// Colour colours the frame glyph; the zero (empty) Colour emits no escape
	// codes.
	Colour ansi.Colour

	// Label is optional postfix text; the empty string omits it. LabelWidth,
	// when > 0, truncates the label to that many columns via [format.Truncate].
	Label      string
	LabelWidth int
}

// frames returns the configured frames, defaulting to [LineFrames].
func (s *Spinner) frames() []string {
	if len(s.Frames) == 0 {
		return LineFrames
	}
	return s.Frames
}

// Tick advances Frame to the next cell, wrapping back to the first at the end.
func (s *Spinner) Tick() {
	s.Frame = (s.Frame + 1) % len(s.frames())
}

// Render draws the current frame glyph, optionally followed by the label.
func (s *Spinner) Render() string {
	frames := s.frames()
	index := s.Frame % len(frames)
	if index < 0 {
		index += len(frames)
	}
	glyph := paint(s.Colour, frames[index])
	if s.Label == "" {
		return glyph
	}
	label := s.Label
	if s.LabelWidth > 0 {
		label = format.Truncate(label, s.LabelWidth)
	}
	return glyph + " " + label
}

var _ Renderer = (*Spinner)(nil)

// Spinner presets, ready to copy and adjust.
var (
	// LineSpinner is a portable ASCII spinner.
	LineSpinner = Spinner{Frames: LineFrames}

	// DotSpinner is a cyan braille spinner.
	DotSpinner = Spinner{Frames: DotFrames, Colour: ansi.Cyan}

	// CircleSpinner is a quartered-circle spinner.
	CircleSpinner = Spinner{Frames: CircleFrames}

	// MoonSpinner is a moon-phase spinner.
	MoonSpinner = Spinner{Frames: MoonFrames}

	// ArrowSpinner is a yellow rotating-arrow spinner.
	ArrowSpinner = Spinner{Frames: ArrowFrames, Colour: ansi.Yellow}

	// PulseSpinner is a magenta growing-and-shrinking bar spinner.
	PulseSpinner = Spinner{Frames: PulseFrames, Colour: ansi.Magenta}

	// SquareSpinner is a quartered-square spinner.
	SquareSpinner = Spinner{Frames: SquareFrames}
)
