package usage

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/spf13/cobra"
)

// view is the resolved model rendered by usage.tmpl. It holds plain strings so
// the template's field names cannot collide with the method set of
// [cobra.Command] (notably Usage, which executes cobra's default usage func as
// a side effect).
type view struct {
	Name        string
	Usage       string
	CommandPath string
}

// newView builds the usage [view] for cmd.
func newView(cmd *cobra.Command) view {
	return view{
		Name:        cmd.Name(),
		Usage:       usageLineOf(cmd),
		CommandPath: cmd.CommandPath(),
	}
}

// usageLineOf returns the usage line for cmd, always terminated with a "[flags]"
// placeholder.
func usageLineOf(cmd *cobra.Command) string {
	use := cmd.Use
	if cmd.HasParent() {
		use = cmd.Parent().CommandPath() + " " + cmd.Use
	}
	if !strings.Contains(use, "[flags]") {
		use += " [flags]"
	}
	return use
}

// Renderer writes the short usage advisory for a [cobra.Command]. The output
// carries richtext styling tags; a richtext writer decides whether they render
// as colour.
type Renderer struct{}

// Render writes the usage advisory for cmd to w. It reports any error from
// writing to w.
func (r Renderer) Render(w io.Writer, cmd *cobra.Command) error {
	tmpl := template.Must(template.New("usage").
		Funcs(tmplfuncs.NewFunc()).
		ParseFS(templateFS, "templates/*.tmpl"))

	// A static template over a well-formed view cannot fail to execute into an
	// in-memory buffer, so a failure here is a template bug, handled like
	// [template.Must]. The only recoverable error is writing to w.
	var buf bytes.Buffer
	template.Must(tmpl, tmpl.ExecuteTemplate(&buf, "usage.tmpl", newView(cmd)))

	body := strings.TrimRight(buf.String(), "\n") + "\n"
	_, err := io.WriteString(w, body)
	return err
}
