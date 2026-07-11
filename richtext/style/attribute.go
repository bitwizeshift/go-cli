package style

import (
	"encoding"
	"errors"
	"fmt"
)

// ErrUnknownAttribute reports that a text-attribute name could not be resolved.
var ErrUnknownAttribute = errors.New("unknown attribute")

// Attribute is a bitfield of text attributes. The zero value is no attributes,
// and values combine with bitwise OR.
type Attribute uint

const (
	Bold Attribute = 1 << iota
	Faint
	Italic
	Underline
	Blink
	Reverse
	Strike
)

// attributeCode pairs an attribute bit with its SGR parameter. The parameters
// are not contiguous (reverse is 7, strike is 9), so the mapping is explicit.
var attributeCodes = []struct {
	attr Attribute
	code int
}{
	{Bold, 1},
	{Faint, 2},
	{Italic, 3},
	{Underline, 4},
	{Blink, 5},
	{Reverse, 7},
	{Strike, 9},
}

var attributeNames = map[string]Attribute{
	"bold":      Bold,
	"faint":     Faint,
	"italic":    Italic,
	"underline": Underline,
	"blink":     Blink,
	"reverse":   Reverse,
	"strike":    Strike,
}

// AttributeByName resolves a single named attribute ("bold", "italic"). The
// second result reports whether the name was recognised.
func AttributeByName(name string) (Attribute, bool) {
	a, ok := attributeNames[name]
	return a, ok
}

// UnmarshalText resolves a single named attribute, returning
// [ErrUnknownAttribute] for an unrecognised name.
func (a *Attribute) UnmarshalText(b []byte) error {
	got, ok := AttributeByName(string(b))
	if !ok {
		return fmt.Errorf("attribute: %w %q", ErrUnknownAttribute, b)
	}
	*a = got
	return nil
}

// params returns the SGR parameters for every attribute set on a, in ascending
// parameter order.
func (a Attribute) params() []int {
	var out []int
	for _, ac := range attributeCodes {
		if a&ac.attr != 0 {
			out = append(out, ac.code)
		}
	}
	return out
}

var _ encoding.TextUnmarshaler = (*Attribute)(nil)
