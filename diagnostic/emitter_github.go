package diagnostic

import (
	"fmt"
	"io"
	"strings"
)

// githubEmitter is an [Emitter] that emits diagnostics in a format that can be
// consumed by GitHub Actions for annotations.
type githubEmitter struct {
	writer io.Writer
}

// Emit emits a diagnostic message in a format that can be consumed by GitHub
// Actions for annotations.
func (e *githubEmitter) Emit(msg *Diagnostic) error {
	fmt.Fprintf(e.writer, "::%s %s::%s\n",
		msg.Severity,
		strings.Join(e.fields(msg), ","),
		e.escapeMessage(msg.Message),
	)
	return nil
}

var _ Emitter = (*githubEmitter)(nil)

// fields returns a list of fields that should be included in the GitHub
// Actions annotation. This includes the title, file, start and end positions.
func (e *githubEmitter) fields(msg *Diagnostic) []string {
	var fields []string
	if msg.Title != "" {
		title := e.escapeField(msg.Title)
		if msg.Code != "" {
			fields = append(fields, fmt.Sprintf("title=[%s] %s", msg.Code, title))
		} else {
			fields = append(fields, fmt.Sprintf("title=%s", title))
		}
	}
	if msg.File != "" {
		fields = append(fields, fmt.Sprintf("file=%s", msg.File))
	}
	if start := msg.Start; start != nil {
		if start.Column > 0 {
			fields = append(fields, fmt.Sprintf("col=%d", start.Column))
		}
		if start.Line > 0 {
			fields = append(fields, fmt.Sprintf("line=%d", start.Line))
		}
	}
	if end := msg.End; end != nil {
		if end.Column > 0 {
			fields = append(fields, fmt.Sprintf("colEnd=%d", end.Column))
		}
		if end.Line > 0 {
			fields = append(fields, fmt.Sprintf("lineEnd=%d", msg.End.Line))
		}
	}
	return fields
}

// escapeMessage escapes characters that don't properly print in GitHub to their
// URL-encoded equivalents.
func (e *githubEmitter) escapeMessage(s string) string {
	return contentEscaper.Replace(s)
}

// escapeField escapes characters that don't properly print in GitHub to their
// URL-encoded equivalents.
func (e *githubEmitter) escapeField(s string) string {
	return fieldEscaper.Replace(s)
}

// contentEscaper is used to encode characters that don't properly print in GitHub
// to their URL-encoded equivalents. This does not need to apply to all
// characters, only those that are problematic -- mostly newlines.
var (
	fieldEscaper = strings.NewReplacer(
		",", "%2C",
	)
	contentEscaper = strings.NewReplacer(
		"\n", "%0A",
		"\r", "%0D",
	)
)
