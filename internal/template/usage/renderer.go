package usage

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/spf13/cobra"
)

// Renderer writes the short usage advisory for a [cobra.Command].
type Renderer struct {
	// Palette styles the output. A nil Palette produces plain output, equivalent
	// to [palette.NoColour].
	Palette palette.Palette
}

// view is the resolved model for the usage advisory.
type view struct {
	Path string
}

// Render writes the usage advisory for cmd to w. It reports any error from
// writing to w.
func (r Renderer) Render(w io.Writer, cmd *cobra.Command) error {
	p := r.Palette
	if p == nil {
		p = palette.NoColour{}
	}
	tmpl := template.Must(template.New("usage").
		Funcs(tmplfuncs.NewFunc(p)).
		ParseFS(templateFS, "templates/*.tmpl"))

	// A static template over a well-formed view cannot fail to execute into an
	// in-memory buffer, so a failure here is a template bug, handled like
	// [template.Must]. The only recoverable error is writing to w.
	var buf bytes.Buffer
	template.Must(tmpl, tmpl.ExecuteTemplate(&buf, "usage.tmpl", view{Path: cmd.CommandPath()}))

	body := strings.TrimRight(buf.String(), "\n") + "\n"
	_, err := io.WriteString(w, body)
	return err
}
