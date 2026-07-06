package spec

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func init() {
	cobra.AddTemplateFuncs(template.DefaultRenderEngine.VersionFuncs())
}

// Build decodes an [Application] specification from r and constructs the
// corresponding [github.com/spf13/cobra.Command] tree, binding each runner to
// the command whose id matches its key.
//
// It returns [ErrUnboundRunner] if a runner is bound to an id with no matching
// command, or a decoding error if r does not hold a valid specification.
func Build(r io.Reader, runners map[string]Runner) (*cobra.Command, error) {
	var app Application
	if err := yaml.NewDecoder(r).Decode(&app); err != nil {
		return nil, err
	}

	unbound := make(map[string]Runner, len(runners))
	for id, runner := range runners {
		unbound[id] = runner
	}
	cmd := app.toCobraCommand(unbound)
	if len(unbound) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnboundRunner, strings.Join(sortedKeys(unbound), ", "))
	}
	annotation.AddIssueURL(cmd, app.IssueURL)
	return cmd, nil
}

// sortedKeys returns the keys of runners in sorted order.
func sortedKeys(runners map[string]Runner) []string {
	keys := make([]string, 0, len(runners))
	for id := range runners {
		keys = append(keys, id)
	}
	slices.Sort(keys)
	return keys
}

// toCobraCommand converts the command info into a [github.com/spf13/cobra.Command],
// removing each bound runner from runners as it is consumed.
func (i *CommandInfo) toCobraCommand(runners map[string]Runner) *cobra.Command {
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
	if runner := runners[i.ID]; runner != nil {
		delete(runners, i.ID)
		cmd.RunE = i.run(runner)
		flag.Register(flag.NewRegistry(cmd.Flags()), runner)
		annotation.ConfigureFlags(cmd)
		annotation.RegisterFlagCompletions(cmd)
	} else {
		cmd.RunE = i.showHelp
	}
	cmd.SetHelpFunc(template.DefaultRenderEngine.HelpFunc())
	cmd.SetUsageFunc(template.DefaultRenderEngine.UsageFunc())
	cmd.SetVersionTemplate(template.DefaultRenderEngine.VersionTemplate())

	for _, group := range i.Commands {
		i.addGroup(cmd, group, runners)
	}
	return cmd
}

// showHelp is the default action for a command with no bound runner. Printing
// the command's help keeps the command runnable, and therefore visible in help
// listings, rather than an inert node cobra hides.
func (*CommandInfo) showHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// addGroup adds the commands of group to cmd. A group named [DefaultGroup] is
// left ungrouped; any other group is registered as a titled cobra group.
func (i *CommandInfo) addGroup(cmd *cobra.Command, group GroupCommandInfo, runners map[string]Runner) {
	groupID := ""
	if group.Name != DefaultGroup {
		groupID = strings.ReplaceAll(group.Name, " ", "-")
		cmd.AddGroup(&cobra.Group{
			ID:    groupID,
			Title: group.Name,
		})
	}
	for _, c := range group.Commands {
		command := c.toCobraCommand(runners)
		command.GroupID = groupID
		cmd.AddCommand(command)
	}
}
