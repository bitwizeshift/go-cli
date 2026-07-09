package progress

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/term/ansi"
)

// defaultWidth is the whole-bar column budget used when [Bar.Width] is unset.
const defaultWidth = 20

// Bar is a renderable, updatable progress bar. Its zero value renders a plain
// bracketless ASCII bar; set fields to customise glyphs, colour, width, and
// postfix text.
type Bar struct {
	// Total is the value representing a full bar; Current is progress toward it.
	Total, Current int64

	// Width is the column budget for the whole bar, decorations included. When
	// <= 0 a default width is used.
	Width int

	// Left and Right bracket the track and are rendered verbatim. Empty strings
	// omit the bracket.
	Left, Right string
	// Empty is the glyph for the unfilled remainder; the empty string defaults
	// to a space.
	Empty string
	// Fill lists cell glyphs from least- to most-filled; the last is a full
	// cell. More than one entry enables sub-cell (fractional) rendering at the
	// frontier. An empty Fill defaults to a single "#".
	Fill []string
	// Head, when non-empty, is drawn as a distinct frontier glyph between the
	// filled and empty regions (for example ">" in "===>"), suppressing sub-cell
	// rendering.
	Head string
	// FillColour and EmptyColour colour the filled and empty regions; the zero
	// (empty) Colour emits no escape codes.
	FillColour, EmptyColour ansi.Colour

	// ShowPercent appends " NN%" after the track when true.
	ShowPercent bool
	// Suffix is optional postfix text; the empty string omits it. SuffixWidth,
	// when > 0, truncates the suffix to that many columns via [format.Truncate].
	Suffix      string
	SuffixWidth int
}

// Fraction returns Current/Total clamped to the range [0,1]. It is 0 when Total
// is not positive.
func (b *Bar) Fraction() float64 {
	if b.Total <= 0 {
		return 0
	}
	switch {
	case b.Current <= 0:
		return 0
	case b.Current >= b.Total:
		return 1
	default:
		return float64(b.Current) / float64(b.Total)
	}
}

// Add increments Current by delta, a convenience over mutating the field
// directly.
func (b *Bar) Add(delta int64) {
	b.Current += delta
}

// Render lays out Left, the filled and empty track, and Right, optionally
// followed by a percentage and the suffix. The track occupies whatever columns
// remain after decorations; a fractional frontier is drawn with [Bar.Head] when
// set, otherwise with a sub-cell glyph from [Bar.Fill].
func (b *Bar) Render() string {
	frac := b.Fraction()
	trailer := b.trailer(frac)
	track := b.track(b.trackWidth(trailer))
	return b.Left + track + b.Right + trailer
}

// trailer is the percentage and suffix rendered after the track.
func (b *Bar) trailer(frac float64) string {
	var sb strings.Builder
	if b.ShowPercent {
		fmt.Fprintf(&sb, " %d%%", int(frac*100))
	}
	if b.Suffix != "" {
		suffix := b.Suffix
		if b.SuffixWidth > 0 {
			suffix = format.Truncate(suffix, b.SuffixWidth)
		}
		sb.WriteString(" " + suffix)
	}
	return sb.String()
}

// trackWidth is the number of track cells available after subtracting the
// brackets and trailer from the configured width.
func (b *Bar) trackWidth(trailer string) int {
	width := b.Width
	if width <= 0 {
		width = defaultWidth
	}
	remaining := width - utf8.RuneCountInString(b.Left) -
		utf8.RuneCountInString(b.Right) - utf8.RuneCountInString(trailer)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// track renders the filled and empty regions across cells, painted with their
// respective colours.
func (b *Bar) track(cells int) string {
	fill := b.Fill
	if len(fill) == 0 {
		fill = []string{"#"}
	}
	empty := b.Empty
	if empty == "" {
		empty = " "
	}

	filled, used := b.filled(cells, fill)
	remainder := strings.Repeat(empty, cells-used)
	return paint(b.FillColour, filled) + paint(b.EmptyColour, remainder)
}

// filled renders the filled portion of the track and reports how many cells it
// consumed, including any frontier cell.
func (b *Bar) filled(cells int, fill []string) (string, int) {
	exact := b.Fraction() * float64(cells)
	full := int(exact)
	if full >= cells {
		return strings.Repeat(fill[len(fill)-1], cells), cells
	}

	var sb strings.Builder
	sb.WriteString(strings.Repeat(fill[len(fill)-1], full))
	if frontier := b.frontier(exact-float64(full), fill); frontier != "" {
		sb.WriteString(frontier)
		return sb.String(), full + 1
	}
	return sb.String(), full
}

// frontier returns the glyph drawn at the boundary between the filled and empty
// regions for a partial cell whose fill level is remainder in [0,1), or "" when
// no frontier cell should be drawn.
func (b *Bar) frontier(remainder float64, fill []string) string {
	if b.Head != "" {
		return b.Head
	}
	if remainder <= 0 || len(fill) < 2 {
		return ""
	}
	// remainder is in [0,1), so index stays within fill.
	index := int(remainder * float64(len(fill)))
	return fill[index]
}

// paint wraps s in colour, resetting afterwards. It leaves s untouched when the
// colour or string is empty.
func paint(colour ansi.Colour, s string) string {
	if colour == "" || s == "" {
		return s
	}
	return string(colour) + s + string(ansi.Reset)
}

// Bar presets, ready to copy and adjust.
var (
	// ASCIIBar is a portable "[#### ]" bar.
	ASCIIBar = Bar{Left: "[", Right: "]", Fill: []string{"#"}, Empty: " ", Width: 30}

	// ArrowBar is a "[===> ]" bar with a distinct head.
	ArrowBar = Bar{Left: "[", Right: "]", Fill: []string{"="}, Head: ">", Empty: " ", Width: 30}

	// ShadedBar fills with a solid block over a shaded remainder ("▓▓▓░░░").
	ShadedBar = Bar{Fill: []string{"▓"}, Empty: "░", Width: 30}

	// BlockBar uses Unicode block glyphs for a smooth, sub-cell green fill.
	BlockBar = Bar{
		Fill:       []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
		Empty:      " ",
		Width:      30,
		FillColour: ansi.Green,
	}

	// SmoothBar fills smooth sub-cell blocks over a shaded background, so the bar
	// appears to grow across a textured track rather than empty space.
	SmoothBar = Bar{
		Fill:        []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
		Empty:       "▒",
		Width:       30,
		FillColour:  ansi.Cyan,
		EmptyColour: ansi.BrightBlack,
	}

	// LineBar draws a heavy line growing over a light one.
	LineBar = Bar{Fill: []string{"━"}, Empty: "─", Width: 30, FillColour: ansi.Green}

	// DotBar fills round markers over hollow ones.
	DotBar = Bar{Fill: []string{"●"}, Empty: "○", Width: 30}
)

var _ Renderer = (*Bar)(nil)
