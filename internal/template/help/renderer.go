package help

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/spf13/cobra"
)

// Renderer writes group-organised help output for a [cobra.Command]. The output
// carries richtext styling tags; a richtext writer decides whether they render
// as colour.
type Renderer struct {
	// Columns is the width that prose, usage, and flag or command descriptions
	// wrap to.
	Columns int

	// Notice, when non-nil, appends an update advisory to the foot of the output.
	Notice *Notice
}

// Render writes the formatted help for cmd to w, listing the positional
// arguments registered on cl. It reports any error from writing to w.
func (r Renderer) Render(w io.Writer, cmd *cobra.Command, cl *arg.CommandLine) error {
	view := NewView(cmd, cl)
	data := struct {
		View
		Columns int
		Notice  *Notice
	}{View: view, Columns: r.Columns, Notice: r.Notice}
	tmpl := template.Must(template.New("help").
		Funcs(funcs(r.Columns, view)).
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
