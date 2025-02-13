package ansi

import (
	"fmt"
)

// SGRControlSequence represents a collection of ANSI SGR attributes.
type SGRControlSequence []SGRParam

// Append the 'other' control sequence to the end of this control sequence.
func (cs *SGRControlSequence) Append(other SGRControlSequence) {
	if *cs == nil {
		*cs = make(SGRControlSequence, 0, len(other))
	}
	*cs = append(*cs, other...)
}

// Format returns the ANSI SGR escape sequence for the format.
func (cs SGRControlSequence) Format(w fmt.State, verb rune) {
	switch verb {
	case 'v':
		_, _ = fmt.Fprint(w, SGRPrefix)
		for i, a := range cs {
			if i > 0 {
				_, _ = fmt.Fprint(w, ";")
			}
			_, _ = fmt.Fprintf(w, "%d", uint8(a))
		}
		_, _ = fmt.Fprint(w, SGRSuffix)
	case 'd':
		for i, a := range cs {
			if i > 0 {
				_, _ = fmt.Fprint(w, ",")
			}
			_, _ = fmt.Fprintf(w, "%d", a)
		}
	default:
		_, _ = fmt.Fprintf(w, "!%c(SGRControlSequence=%d)", verb, cs)
	}
}

// String returns the ANSI SGR escape sequence for the format.
func (cs SGRControlSequence) String() string {
	return fmt.Sprintf("%v", cs)
}

var (
	_ fmt.Formatter = (*SGRControlSequence)(nil)
	_ fmt.Stringer  = (*SGRControlSequence)(nil)
)
