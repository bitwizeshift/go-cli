package panichandler

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
)

// Renderer writes a panic report describing an uncaught panic.
type Renderer struct {
	// Palette styles the output. A nil Palette produces plain output, equivalent
	// to [palette.NoColour].
	Palette palette.Palette
}

// Render writes the panic report for ctx to w. It reports any error from writing
// to w.
func (r Renderer) Render(w io.Writer, ctx PanicContext) error {
	p := r.Palette
	if p == nil {
		p = palette.NoColour{}
	}
	tmpl := template.Must(template.New("panic").
		Funcs(funcs(p)).
		ParseFS(templateFS, "templates/*.tmpl"))

	// A static template over well-formed data cannot fail to execute into an
	// in-memory buffer, so a failure here is a template bug, handled like
	// [template.Must]. The only recoverable error is writing to w.
	var buf bytes.Buffer
	template.Must(tmpl, tmpl.ExecuteTemplate(&buf, "panic.tmpl", newData(ctx)))

	body := strings.TrimRight(buf.String(), "\n") + "\n"
	_, err := io.WriteString(w, body)
	return err
}
