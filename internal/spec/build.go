package spec

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/arity"
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

const (
	// idSeparator delimits the names composing a command's id path.
	idSeparator = "."

	// commandOperand names the subcommand that must be chosen to reach a runner,
	// as shown in the usage of a command that has subcommands.
	commandOperand = "<command>"
)

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
	// Builders binds each command id path to the runner that executes it. An id
	// path is the dot-delimited names of a command and its ancestors, so the
	// command invoked as "app remote add" is bound as "app.remote.add".
	Builders map[string]Builder

	// Theme resolves the styling tags emitted by the output templates. A nil
	// Theme uses [richtext.DefaultTheme].
	Theme *richtext.Theme

	// Colour selects the colour policy applied to the wrapped output streams.
	Colour ColourMode

	// Version is the running build's version, reported by the root command.
	Version string

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
// the command whose id path matches its key. Every command writes styled output
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
	cmd, cl := app.toCobraCommand(app.Name, unbound, store)
	cmd.Version = opts.Version
	if len(unbound) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnboundRunner, strings.Join(sortedKeys(unbound), ", "))
	}
	argdef.AddIssueURL(cmd, app.IssueURL)
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
// removing each bound runner from runners as it is consumed. path is the
// dot-delimited id path identifying this command, and is the key a runner is
// bound under. store is shared by every command so a bound runner can reach the
// application's storage roots. It returns the command alongside its argument cl,
// which is nil when no runner is bound.
func (i *CommandInfo) toCobraCommand(path string, builders map[string]Builder, store *storage.AppStorage) (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{
		Short:         i.Summary,
		Long:          i.Description,
		Example:       strings.Join(i.Examples, "\n"),
		Aliases:       i.Aliases,
		Hidden:        i.Hidden,
		Deprecated:    i.Deprecated,
		SilenceUsage:  true,
		SilenceErrors: true,

		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},

		SuggestionsMinimumDistance: 1,
	}
	var cl *arg.CommandLine
	if builder := builders[path]; builder != nil {
		delete(builders, path)
		cl = (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
		arg.Register(cl, builder)
		argdef.VerifyPositionals((*argdef.CommandLine)(cl))
		cmd.Args = positionalArgs(argdef.Arity((*argdef.CommandLine)(cl)))
		argdef.ConfigureFlags(cmd)
		completion.RegisterFlags(cmd)
		cmd.ValidArgsFunction = completion.ForArgs(
			argdef.PositionalCompletions((*argdef.CommandLine)(cl)),
			argdef.UnmatchedCompletion((*argdef.CommandLine)(cl)),
		)
		cmd.RunE = i.run(builder, store, cl)
	} else {
		cmd.Args = positionalArgs(argdef.Arity(argdef.New()))
		cmd.RunE = i.showHelp
	}
	cmd.SetHelpFunc(template.DefaultRenderEngine.HelpFunc(cl))
	cmd.SetUsageFunc(template.DefaultRenderEngine.UsageFunc())
	cmd.SetVersionTemplate(template.DefaultRenderEngine.VersionTemplate())

	for _, group := range i.Commands {
		i.addGroup(cmd, path, group, builders, store)
	}
	cmd.Use = i.usage(cmd, cl)
	return cmd, cl
}

// usage returns the synopsis cobra displays for cmd: the command's name
// followed by the arguments registered on cl. A command with subcommands names
// one as an operand, since a subcommand must be chosen to reach a runner.
func (i *CommandInfo) usage(cmd *cobra.Command, cl *arg.CommandLine) string {
	if cl == nil {
		cl = (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	}
	var operands []string
	if cmd.HasAvailableSubCommands() {
		operands = append(operands, commandOperand)
	}
	usage := argdef.Usage((*argdef.CommandLine)(cl), operands...)
	if usage == "" {
		return i.Name
	}
	return i.Name + " " + usage
}

// positionalArgs adapts a permitted argument count into cobra's
// positional-argument validator.
func positionalArgs(a arity.Arity) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		return a.Validate(len(args))
	}
}

// showHelp is the default action for a command with no bound runner. Printing
// the command's help keeps the command runnable, and therefore visible in help
// listings, rather than an inert node cobra hides.
func (*CommandInfo) showHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// addGroup adds the commands of group to cmd, each identified by its name
// appended to path. A group named [DefaultGroup] is left ungrouped; any other
// group is registered as a titled cobra group.
func (i *CommandInfo) addGroup(cmd *cobra.Command, path string, group GroupCommandInfo, builders map[string]Builder, store *storage.AppStorage) {
	groupID := ""
	if group.Name != DefaultGroup {
		groupID = strings.ReplaceAll(group.Name, " ", "-")
		cmd.AddGroup(&cobra.Group{
			ID:    groupID,
			Title: group.Name,
		})
	}
	for _, c := range group.Commands {
		command, _ := c.toCobraCommand(path+idSeparator+c.Name, builders, store)
		command.GroupID = groupID
		cmd.AddCommand(command)
	}
}
