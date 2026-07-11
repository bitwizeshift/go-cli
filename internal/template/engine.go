package template

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/template/tag"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/internal/template/usage"
	"github.com/bitwizeshift/go-cli/internal/template/version"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/spf13/cobra"
)

// RenderEngine coordinates construction of the underlying Renderer objects.
//
// The renderers emit richtext styling tags rather than colour; whether those
// tags render as colour is decided by the richtext writer wrapping the
// destination. The engine therefore only sizes output to the terminal.
type RenderEngine struct {
	Sizer term.Sizer
}

// DefaultRenderEngine is the standard configuration for the render engine.
var DefaultRenderEngine = RenderEngine{
	Sizer: term.DefaultSizer,
}

// HelpRenderer returns the help renderer, sized for the terminal behind w.
func (re RenderEngine) HelpRenderer(w io.Writer) *help.Renderer {
	return &help.Renderer{Columns: re.Sizer.Columns(baseWriter(w))}
}

// HelpFunc returns a cobra func that can be installed with
// cobra.Command.SetHelpFunc
func (re RenderEngine) HelpFunc() func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		stdout := cmd.OutOrStdout()
		_ = re.HelpRenderer(stdout).Render(stdout, cmd)
	}
}

// UsageRenderer returns the usage renderer.
func (re RenderEngine) UsageRenderer() *usage.Renderer {
	return &usage.Renderer{}
}

// UsageFunc returns a cobra func that can be installed with
// cobra.Command.SetUsageFunc
func (re RenderEngine) UsageFunc() func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		return re.UsageRenderer().Render(cmd.ErrOrStderr(), cmd)
	}
}

// PanicRenderer returns the panic-report renderer.
func (re RenderEngine) PanicRenderer() *panichandler.Renderer {
	return &panichandler.Renderer{}
}

// VersionTemplate returns the --version template text for
// cobra.Command.SetVersionTemplate.
func (re RenderEngine) VersionTemplate() string {
	return version.Template()
}

// VersionFuncs returns the template functions the version template requires, for
// registration with cobra.AddTemplateFuncs.
func (re RenderEngine) VersionFuncs() template.FuncMap {
	return tmplfuncs.NewFunc()
}

func (re RenderEngine) Errorf(w io.Writer, f string, args ...any) error {
	const prefix = "error: "
	spacePrefix := strings.Repeat(" ", len(prefix))

	columns := re.Sizer.Columns(baseWriter(w))
	message := fmt.Sprintf(f, args...)
	message = format.Resize(message, columns-len(prefix))
	lines := strings.Split(message, "\n")

	var sb strings.Builder
	_, _ = sb.WriteString(tag.Themed("error", prefix))
	_, _ = sb.WriteString(tag.Raw(lines[0]))
	for _, line := range lines[1:] {
		_, _ = sb.WriteString("\n")
		_, _ = sb.WriteString(spacePrefix)
		_, _ = sb.WriteString(tag.Raw(line))
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}

func (re RenderEngine) FlagErrorFunc() func(cmd *cobra.Command, err error) error {
	return func(cmd *cobra.Command, err error) error {
		return re.Errorf(cmd.ErrOrStderr(), "%v", err)
	}
}

// baseWriter unwraps w through any writers that expose their destination via a
// Writer() method, returning the base stream. It lets terminal introspection
// (such as sizing) see past a wrapping richtext writer to the real stream.
func baseWriter(w io.Writer) io.Writer {
	for {
		unwrapper, ok := w.(interface{ Writer() io.Writer })
		if !ok {
			return w
		}
		w = unwrapper.Writer()
	}
}
