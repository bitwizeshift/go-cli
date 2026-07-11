package panichandler

import (
	"bytes"
	"io"
	"strings"
	"text/template"
)

// Renderer writes a panic report describing an uncaught panic. The output
// carries richtext styling tags; a richtext writer decides whether they render
// as colour.
type Renderer struct{}

// Render writes the panic report for ctx to w. It reports any error from writing
// to w.
func (r Renderer) Render(w io.Writer, ctx PanicContext) error {
	tmpl := template.Must(template.New("panic").
		Funcs(funcs()).
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
