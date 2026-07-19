package spec

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/completion"
	"github.com/bitwizeshift/go-cli/internal/storage"
	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func init() {
	cobra.AddTemplateFuncs(template.DefaultRenderEngine.VersionFuncs())
}

// ColourMode selects how colour is decided for a command tree's output streams.
type ColourMode int

const (
	// ColourAuto decides colour from the destination terminal.
	ColourAuto ColourMode = iota
	// ColourDisabled never emits colour.
	ColourDisabled
	// ColourEnabled always emits colour.
	ColourEnabled
)

// Options configures how a command tree is built and how its output is styled.
type Options struct {
	// Builders binds each command id to the runner that executes it.
	Builders map[string]Builder

	// Theme resolves the styling tags emitted by the output templates. A nil
	// Theme uses [richtext.DefaultTheme].
	Theme *richtext.Theme

	// Colour selects the colour policy applied to the wrapped output streams.
	Colour ColourMode

	// Stdout and Stderr are the base streams wrapped for styled output. A nil
	// value uses [os.Stdout] or [os.Stderr] respectively.
	Stdout io.Writer
	Stderr io.Writer

	// Update configures update-availability checking. Checking is enabled only
	// when it carries a version, a source, and at least one provider.
	Update UpdateOptions
}

// Build decodes an [Application] specification from r and constructs the
// corresponding [github.com/spf13/cobra.Command] tree, binding each runner to
// the command whose id matches its key. Every command writes styled output
// through a [richtext.Writer] wrapping the configured streams; the writers must
// be flushed by [Execute] once the tree has run.
//
// It returns [ErrUnboundRunner] if a runner is bound to an id with no matching
// command, or a decoding error if r does not hold a valid specification.
func Build(r io.Reader, opts Options) (*cobra.Command, error) {
	var app Application
	if err := yaml.NewDecoder(r).Decode(&app); err != nil {
		return nil, err
	}

	unbound := make(map[string]Builder, len(opts.Builders))
	maps.Copy(unbound, opts.Builders)
	store := storage.NewAppStorage(app.resolveAppID())
	cmd, cl := app.toCobraCommand(unbound, store)
	if len(unbound) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnboundRunner, strings.Join(sortedKeys(unbound), ", "))
	}
	annotation.AddIssueURL(cmd, app.IssueURL)
	checker, err := opts.Update.checker(&app, store.Cache)
	if err != nil {
		return nil, err
	}
	if checker != nil {
		installUpdateHelp(cmd, checker, cl)
	}
	setStreams(cmd,
		opts.newWriter(opts.Stdout, os.Stdout),
		opts.newWriter(opts.Stderr, os.Stderr),
	)
	return cmd, nil
}

// newWriter wraps base (or fallback when base is nil) in a [richtext.Writer]
// configured for the options' theme and colour policy.
func (o Options) newWriter(base, fallback io.Writer) *richtext.Writer {
	if base == nil {
		base = fallback
	}
	theme := o.Theme
	if theme == nil {
		theme = richtext.DefaultTheme
	}
	w := richtext.NewWriter(base, theme)
	switch o.Colour {
	case ColourDisabled:
		w.EnableColour(false)
	case ColourEnabled:
		w.ForceColour()
	}
	return w
}

// setStreams routes cmd and every command beneath it through out and err.
func setStreams(cmd *cobra.Command, out, err io.Writer) {
	cmd.SetOut(out)
	cmd.SetErr(err)
	for _, sub := range cmd.Commands() {
		setStreams(sub, out, err)
	}
}

// sortedKeys returns the keys of runners in sorted order.
func sortedKeys(builders map[string]Builder) []string {
	keys := make([]string, 0, len(builders))
	for id := range builders {
		keys = append(keys, id)
	}
	slices.Sort(keys)
	return keys
}

// toCobraCommand converts the command info into a [github.com/spf13/cobra.Command],
// removing each bound runner from runners as it is consumed. store is shared by
// every command so a bound runner can reach the application's storage roots. It
// returns the command alongside its argument cl, which is nil when no
// runner is bound.
func (i *CommandInfo) toCobraCommand(builders map[string]Builder, store *storage.AppStorage) (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{
		Use:           i.Use,
		Short:         i.Summary,
		Long:          i.Description,
		Example:       strings.Join(i.Examples, "\n"),
		Version:       i.Version,
		Aliases:       i.Aliases,
		Hidden:        i.Hidden,
		Deprecated:    i.Deprecated,
		Args:          cobra.PositionalArgs(i.Arity),
		SilenceUsage:  true,
		SilenceErrors: true,

		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},

		SuggestionsMinimumDistance: 1,
	}
	var cl *arg.CommandLine
	if builder := builders[i.ID]; builder != nil {
		delete(builders, i.ID)
		cl = (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
		arg.Register(cl, builder)
		annotation.ConfigureFlags(cmd)
		completion.RegisterFlags(cmd)
		cmd.ValidArgsFunction = completion.ForArgs(
			argdef.PositionalCompletions((*argdef.CommandLine)(cl)),
			argdef.UnmatchedCompletion((*argdef.CommandLine)(cl)),
		)
		cmd.RunE = i.run(builder, store, cl)
	} else {
		cmd.RunE = i.showHelp
	}
	cmd.SetHelpFunc(template.DefaultRenderEngine.HelpFunc(cl))
	cmd.SetUsageFunc(template.DefaultRenderEngine.UsageFunc())
	cmd.SetVersionTemplate(template.DefaultRenderEngine.VersionTemplate())

	for _, group := range i.Commands {
		i.addGroup(cmd, group, builders, store)
	}
	return cmd, cl
}

// showHelp is the default action for a command with no bound runner. Printing
// the command's help keeps the command runnable, and therefore visible in help
// listings, rather than an inert node cobra hides.
func (*CommandInfo) showHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// addGroup adds the commands of group to cmd. A group named [DefaultGroup] is
// left ungrouped; any other group is registered as a titled cobra group.
func (i *CommandInfo) addGroup(cmd *cobra.Command, group GroupCommandInfo, builders map[string]Builder, store *storage.AppStorage) {
	groupID := ""
	if group.Name != DefaultGroup {
		groupID = strings.ReplaceAll(group.Name, " ", "-")
		cmd.AddGroup(&cobra.Group{
			ID:    groupID,
			Title: group.Name,
		})
	}
	for _, c := range group.Commands {
		command, _ := c.toCobraCommand(builders, store)
		command.GroupID = groupID
		cmd.AddCommand(command)
	}
}
