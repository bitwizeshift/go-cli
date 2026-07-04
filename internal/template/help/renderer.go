package help

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/spf13/cobra"
)

// Renderer writes group-organised help output for a [cobra.Command].
type Renderer struct {
	// Columns is the width that prose, usage, and flag or command descriptions
	// wrap to.
	Columns int

	// Palette styles the output. A nil Palette produces plain output, equivalent
	// to [palette.NoColour].
	Palette palette.Palette
}

// Render writes the formatted help for cmd to w. It reports any error from
// writing to w.
func (r Renderer) Render(w io.Writer, cmd *cobra.Command) error {
	p := r.Palette
	if p == nil {
		p = palette.NoColour{}
	}
	view := NewView(cmd)
	data := struct {
		View
		Columns int
	}{View: view, Columns: r.Columns}
	tmpl := template.Must(template.New("help").
		Funcs(funcs(r.Columns, p, view)).
		ParseFS(templateFS, "templates/*.tmpl"))

	// A static template over a well-formed [View] cannot fail to execute into an
	// in-memory buffer, so a failure here is a template bug, handled like
	// [template.Must]. The only recoverable error is writing to w.
	var buf bytes.Buffer
	template.Must(tmpl, tmpl.ExecuteTemplate(&buf, "help.tmpl", data))

	body := strings.TrimRight(buf.String(), "\n") + "\n"
	_, err := io.WriteString(w, body)
	return err
}
