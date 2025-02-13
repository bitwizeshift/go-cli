package wrap

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

// String wraps a string to a maximum column width.
func String(s string) string {
	return defaultWrapper.String(s)
}

// Strings wraps a series of strings representing lines to a maximum column width.
// If any of the 'lines' contain newlines, they will be passed through verbatim.
func Strings(lines ...string) string {
	return defaultWrapper.Strings(lines...)
}

// Lines wraps a slice of strings to a maximum column width.
// If any of the 'lines' contain newlines, they will be passed through verbatim.
func Lines(lines ...string) []string {
	return defaultWrapper.Lines(lines...)
}

type Wrapper struct {
	MaxWidth int
}

func FromTerminal(w io.Writer) (*Wrapper, error) {
	if v, ok := w.(interface{ Fd() uint64 }); ok {
		width, _, err := term.GetSize(int(v.Fd()))
		if err != nil {
			return nil, err
		}
		return &Wrapper{
			MaxWidth: width,
		}, nil
	}
	return nil, fmt.Errorf("wrapper: cannot determine terminal width")
}

func (w *Wrapper) String(s string) string {
	return w.Strings(strings.Split(s, "\n")...)
}

func (w *Wrapper) Strings(lines ...string) string {
	return strings.Join(w.Lines(lines...), "\n")
}

func (w *Wrapper) Lines(lines ...string) []string {
	if len(lines) == 0 {
		return nil
	}
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	var newlines []string
	var sb strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			newlines = append(newlines, "")
			continue
		}
		for _, word := range strings.Fields(line) {
			if sb.Len()+len(word)+1 > w.maxWidth() {
				newlines = append(newlines, sb.String())
				sb.Reset()
			}
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(word)
		}
	}
	if sb.Len() > 0 {
		newlines = append(newlines, sb.String())
		sb.Reset()
	}
	return newlines
}

func (w *Wrapper) maxWidth() int {
	if w == nil || w.MaxWidth == 0 {
		return defaultWrapper.MaxWidth
	}
	return w.MaxWidth
}

var defaultWrapper = &Wrapper{
	MaxWidth: 120,
}
