package diagslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/term"
)

// TextHandler renders Diagnostics in a rustc-style human-readable form,
// optionally with ANSI colour.
type TextHandler struct {
	w       io.Writer
	columns int
}

// NewTextHandler returns a TextHandler that writes to w. Colour is enabled
// when enabler returns true for w.
func NewTextHandler(w io.Writer, sizer term.Sizer) *TextHandler {
	return &TextHandler{
		w:       w,
		columns: sizer.Columns(w),
	}
}

// Enabled implements slog.Handler.
func (h *TextHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

// Handle implements slog.Handler.
func (h *TextHandler) Handle(_ context.Context, r slog.Record) error {
	diagnostic := RecordFromSlog(r)

	severity := strings.ToLower(r.Level.String())
	var location string
	location = h.formatLocation(
		diagnostic.File,
		diagnostic.LineStart,
		diagnostic.LineEnd,
		diagnostic.ColumnStart,
		diagnostic.ColumnEnd,
	)

	id := diagnostic.ID
	title := diagnostic.Title
	message := diagnostic.Message

	severityFmt := severityColour(r.Level)
	accentFmt := severityFmt
	idFmt := themeFormat("emphasis")
	linkFmt := themeFormat("url")
	rawFmt := noFormatter{}

	/*
	   Example output:
	   error[E0499]: cannot borrow `x` as mutable more than once
	     --> src/main.rs:10:5-10
	*/
	headline := h.firstNonEmpty(title, message)
	if id != "" {
		header := severityFmt.Format("%s[", severity) + idFmt.Format("%s", rawFmt.Format("%s", id)) + severityFmt.Format("]")
		_, _ = fmt.Fprintf(h.w, "%s: %s\n", header, headline)
	} else {
		_, _ = fmt.Fprintf(h.w, "%s: %s\n", severityFmt.Format("%s", severity), headline)
	}
	if location != "" {
		_, _ = fmt.Fprintf(
			h.w,
			"  %s %s\n",
			accentFmt.Format("-->"),
			linkFmt.Format("%s", location),
		)
	}
	if title != "" && message != "" && title != message {
		// _, _ = fmt.Fprintf(h.w, "   %s\n", accentFmt.Format("|"))
		text := format.Resize(message, h.columns-5)
		for line := range strings.SplitSeq(text, "\n") {
			_, _ = fmt.Fprintf(h.w, "   %s %s\n",
				accentFmt.Format("|"),
				rawFmt.Format("%s", line),
			)
		}
	}
	_, _ = fmt.Fprintln(h.w)
	return nil
}

func severityColour(level slog.Level) themeFormat {
	switch level {
	case slog.LevelError:
		return "error"
	case slog.LevelWarn:
		return "warning"
	case slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

// WithAttrs implements slog.Handler. Attrs are ignored; Diagnostics carry
// their own attributes.
func (h *TextHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

// WithGroup implements slog.Handler. Groups are ignored; Diagnostics carry
// their own attributes.
func (h *TextHandler) WithGroup(string) slog.Handler {
	return h
}

var _ slog.Handler = (*TextHandler)(nil)

func (h *TextHandler) formatLocation(file string, lineStart int64, lineEnd int64, colStart int64, colEnd int64) string {
	var b strings.Builder
	if file != "" {
		b.WriteString(file)
	}
	if lineStart > 0 {
		b.WriteString(":")
		fmt.Fprintf(&b, "%d", lineStart)
		if lineEnd > 0 && lineEnd != lineStart {
			b.WriteString("-")
			fmt.Fprintf(&b, "%d", lineEnd)
		}
		if colStart > 0 {
			b.WriteString(":")
			fmt.Fprintf(&b, "%d", colStart)
			if colEnd > 0 && colEnd != colStart {
				b.WriteString("-")
				fmt.Fprintf(&b, "%d", colEnd)
			}
		}
	}
	return b.String()
}
func (h *TextHandler) firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

type themeFormat string

func (tf themeFormat) Format(format string, args ...any) string {
	return fmt.Sprintf("[theme:%s]%s[/theme]", string(tf), fmt.Sprintf(format, args...))
}

type noFormatter struct{}

func (noFormatter) Format(format string, args ...any) string {
	return fmt.Sprintf("[richtext:off]%s[/richtext]", fmt.Sprintf(format, args...))
}
