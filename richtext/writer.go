package richtext

import (
	"io"

	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/richtext/internal/sgr"
	"github.com/bitwizeshift/go-cli/richtext/internal/token"
	"github.com/bitwizeshift/go-cli/richtext/style"
)

// The recognised tag namespaces. Any other namespace is not a tag and is passed
// through as literal text.
const (
	nsForeground = "fg"
	nsBackground = "bg"
	nsAttribute  = "attr"
	nsTheme      = "theme"
)

// Writer renders bracketed tag markup to an underlying writer as ANSI escapes.
//
// Tags take the form [ns:field] and close with [/ns]. The namespaces are:
//   - fg / bg: a foreground or background colour, either one of the sixteen
//     named ANSI colours or rgb(r,g,b);
//   - attr: a text attribute such as bold or italic, which accumulates while
//     nested;
//   - theme: a named style registered in the [Theme] passed to [NewWriter].
//
// Tags must close in the reverse of the order they were opened. A known
// namespace with an unrecognised field renders as a reset. An unknown namespace
// is not a tag: it and its closing tag are emitted verbatim.
//
// Colour output is governed by a [term.ColourEnabler], [term.DefaultEnabler] by
// default; see [Writer.EnableColour] and [Writer.ForceColour]. When colour is
// disabled, text is written through unchanged and no escapes are emitted.
type Writer struct {
	dst      io.Writer
	enabler  term.ColourEnabler
	theme    *Theme
	scanner  token.Scanner
	stack    []frame
	lastBody string
}

// frame is one open tag on the render stack.
type frame struct {
	namespace string
	reset     bool // a known namespace with an unrecognised field
	colour    style.Colour
	attribute style.Attribute
	themed    style.Style
}

// NewWriter returns a Writer that renders to dst, resolving [theme:name] tags
// against theme. theme may be nil, in which case every theme tag renders as a
// reset. The default colour policy is [term.DefaultEnabler].
func NewWriter(dst io.Writer, theme *Theme) *Writer {
	return &Writer{
		dst:     dst,
		enabler: term.DefaultEnabler,
		theme:   theme,
	}
}

// EnableColour selects the default colour policy when b is true, or disables
// colour entirely when b is false.
func (w *Writer) EnableColour(b bool) {
	if b {
		w.enabler = term.DefaultEnabler
		return
	}
	w.enabler = term.FixedEnabler(false)
}

// ForceColour emits colour unconditionally, regardless of the destination.
func (w *Writer) ForceColour() {
	w.enabler = term.FixedEnabler(true)
}

// Write implements [io.Writer]. It returns a [*TagError] wrapping
// [ErrUnbalancedTag] when a closing tag does not match the open tag.
func (w *Writer) Write(p []byte) (int, error) {
	for _, tok := range w.scanner.Scan(p) {
		if err := w.handle(tok); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

// Close flushes any trailing partial tag as literal text and reports whether
// the markup was balanced. It returns a [*TagError] wrapping [ErrUnclosedTag]
// if any tags remain open. Close does not close the underlying writer.
func (w *Writer) Close() error {
	if tok, ok := w.scanner.Flush(); ok {
		if err := w.writeString(tok.Raw); err != nil {
			return err
		}
	}
	if len(w.stack) > 0 {
		top := w.stack[len(w.stack)-1]
		return &TagError{Namespace: top.namespace, Err: ErrUnclosedTag}
	}
	return nil
}

func (w *Writer) handle(tok token.Token) error {
	switch tok.Kind {
	case token.Open:
		if !isKnownNamespace(tok.Namespace) {
			return w.writeString(tok.Raw)
		}
		w.stack = append(w.stack, w.openFrame(tok.Namespace, tok.Field))
		return w.emit()
	case token.Close:
		if !isKnownNamespace(tok.Namespace) {
			return w.writeString(tok.Raw)
		}
		if len(w.stack) == 0 || w.stack[len(w.stack)-1].namespace != tok.Namespace {
			return &TagError{Namespace: tok.Namespace, Err: ErrUnbalancedTag}
		}
		w.stack = w.stack[:len(w.stack)-1]
		return w.emit()
	default:
		return w.writeString(tok.Raw)
	}
}

// openFrame builds the render frame for an opening tag, marking it as a reset
// when the field is not recognised for its namespace.
func (w *Writer) openFrame(namespace, field string) frame {
	f := frame{namespace: namespace}
	switch namespace {
	case nsForeground, nsBackground:
		var c style.Colour
		if c.UnmarshalText([]byte(field)) != nil {
			f.reset = true
		} else {
			f.colour = c
		}
	case nsAttribute:
		var a style.Attribute
		if a.UnmarshalText([]byte(field)) != nil {
			f.reset = true
		} else {
			f.attribute = a
		}
	case nsTheme:
		if s, ok := w.lookupTheme(field); ok {
			f.themed = s
		} else {
			f.reset = true
		}
	}
	return f
}

func (w *Writer) lookupTheme(name string) (style.Style, bool) {
	if w.theme == nil {
		return style.Style{}, false
	}
	return w.theme.lookup(name)
}

// resolve collapses the render stack into the currently active style.
func (w *Writer) resolve() style.Style {
	var s style.Style
	for _, f := range w.stack {
		switch {
		case f.reset:
			s = style.Style{}
		case f.namespace == nsForeground:
			s.Foreground = f.colour
		case f.namespace == nsBackground:
			s.Background = f.colour
		case f.namespace == nsAttribute:
			s.Attributes |= f.attribute
		case f.namespace == nsTheme:
			s = f.themed
		}
	}
	return s
}

// emit writes the escape needed to move the terminal to the active style,
// skipping output when colour is disabled or the style is unchanged.
func (w *Writer) emit() error {
	if !w.enabler.EnableColour(w.dst) {
		return nil
	}
	body := w.resolve().String()
	if body == w.lastBody {
		return nil
	}
	w.lastBody = body
	if body == "" {
		return w.writeString(sgr.Reset)
	}
	return w.writeString(sgr.Reset + body)
}

func (w *Writer) writeString(s string) error {
	_, err := io.WriteString(w.dst, s)
	return err
}

func isKnownNamespace(namespace string) bool {
	switch namespace {
	case nsForeground, nsBackground, nsAttribute, nsTheme:
		return true
	default:
		return false
	}
}

var _ io.Writer = (*Writer)(nil)
