package template

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/bitwizeshift/go-cli/internal/format"
	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/internal/template/usage"
	"github.com/bitwizeshift/go-cli/internal/template/version"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/spf13/cobra"
)

// RenderEngine coordinates construction of the underlying Renderer objects
type RenderEngine struct {
	ColourEnabler term.ColourEnabler
	Sizer         term.Sizer
}

// DefaultRenderEngine is the standard configuration for the render
// engine
var DefaultRenderEngine = RenderEngine{
	ColourEnabler: term.DefaultEnabler,
	Sizer:         term.DefaultSizer,
}

// HelpRenderer returns the appropriate renderer for the Usage template
// based on the writer that it will be written to.
func (re RenderEngine) HelpRenderer(w io.Writer) *help.Renderer {
	enabled := re.ColourEnabler.EnableColour(w)
	columns := re.Sizer.Columns(w)
	var p palette.Palette = palette.NoColour{}
	if enabled {
		p = palette.DefaultColour
	}
	return &help.Renderer{
		Columns: columns,
		Palette: p,
	}
}

// HelpFunc returns a cobra func that can be installed with
// cobra.Command.SetHelpFunc
func (re RenderEngine) HelpFunc() func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		stdout := cmd.OutOrStdout()
		_ = re.HelpRenderer(stdout).Render(stdout, cmd)
	}
}

// UsageRenderer returns the appropriate renderer for the Usage template
// based on the writer that it will be written to.
func (re RenderEngine) UsageRenderer(w io.Writer) *usage.Renderer {
	enabled := re.ColourEnabler.EnableColour(w)
	var p palette.Palette = palette.NoColour{}
	if enabled {
		p = palette.DefaultColour
	}
	return &usage.Renderer{
		Palette: p,
	}
}

// UsageFunc returns a cobra func that can be installed with
// cobra.Command.SetUsageFunc
func (re RenderEngine) UsageFunc() func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		stderr := cmd.ErrOrStderr()
		return re.UsageRenderer(stderr).Render(stderr, cmd)
	}
}

// PanicRenderer returns the appropriate renderer for the Panic template
// based on the writer that it will be written to.
func (re RenderEngine) PanicRenderer(w io.Writer) *panichandler.Renderer {
	enabled := re.ColourEnabler.EnableColour(w)
	var p palette.Palette = palette.NoColour{}
	if enabled {
		p = palette.DefaultColour
	}
	return &panichandler.Renderer{
		Palette: p,
	}
}

// VersionTemplate returns the --version template text for
// cobra.Command.SetVersionTemplate.
func (re RenderEngine) VersionTemplate() string {
	return version.Template()
}

// VersionFuncs returns the template functions the version template requires,
// for registration with cobra.AddTemplateFuncs. Colour is decided for
// [os.Stdout], where version output is written.
func (re RenderEngine) VersionFuncs() template.FuncMap {
	enabled := re.ColourEnabler.EnableColour(os.Stdout)
	var p palette.Palette = palette.NoColour{}
	if enabled {
		p = palette.DefaultColour
	}
	return tmplfuncs.NewFunc(p)
}

func (re RenderEngine) Errorf(w io.Writer, f string, args ...any) error {
	const prefix = "error: "
	spacePrefix := strings.Repeat(" ", len(prefix))

	enabled := re.ColourEnabler.EnableColour(w)
	var p palette.Palette = palette.NoColour{}
	if enabled {
		p = palette.DefaultColour
	}
	columns := re.Sizer.Columns(w)
	message := fmt.Sprintf(f, args...)
	message = format.Resize(message, columns-len(prefix))
	lines := strings.Split(message, "\n")

	var sb strings.Builder
	_, _ = sb.WriteString(p.Error(prefix))
	_, _ = sb.WriteString(lines[0])
	for _, line := range lines[1:] {
		_, _ = sb.WriteString("\n")
		_, _ = sb.WriteString(spacePrefix)
		_, _ = sb.WriteString(line)
	}
	_, err := w.Write([]byte(sb.String()))
	return err
}

func (re RenderEngine) FlagErrorFunc() func(cmd *cobra.Command, err error) error {
	return func(cmd *cobra.Command, err error) error {
		return re.Errorf(cmd.ErrOrStderr(), "%v", err)
	}
}
