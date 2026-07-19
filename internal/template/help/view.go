package help

import (
	"strings"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/spf13/cobra"
)

// additionalCommands is the group title used for subcommands that do not belong
// to any named cobra group.
const additionalCommands = "Additional Commands"

// View is the resolved help model for a command: its title, description, usage,
// examples, grouped subcommands and flags, and trailing advice. It is what the
// help templates render, and it is exported so its derivation can be tested
// directly.
type View struct {
	Name          string
	Description   string
	Usage         string
	Examples      []string
	CommandGroups []CommandGroup
	Arguments     []ArgumentInfo
	FlagGroups    []FlagGroup
	Hint          Hint
}

// ArgumentInfo is a single argument entry in a help listing. A variadic
// argument claims every remaining argument rather than a single slot.
type ArgumentInfo struct {
	Name     string
	Type     string
	Usage    string
	Required bool
	Variadic bool
}

// CommandGroup is a titled list of subcommands.
type CommandGroup struct {
	Title    string
	Commands []Command
}

// Command is a single subcommand entry in a help listing.
type Command struct {
	Name    string
	Summary string
}

// FlagGroup is a titled list of flags.
type FlagGroup struct {
	Title string
	Flags []FlagInfo
}

// FlagInfo is a single flag entry in a help listing. Type is empty for flags
// that take no value argument (such as booleans).
type FlagInfo struct {
	Shorthand string
	Name      string
	Type      string
	Usage     string
}

// Hint models the trailing "--help" advice shown for commands that have visible
// subcommands. Path is the command path the advice refers to, and is only
// meaningful when Show is true.
type Hint struct {
	Show bool
	Path string
}

// Notice models the trailing advisory shown when a newer release of the
// application is available. Current and Latest are the running and available
// versions.
type Notice struct {
	Current string
	Latest  string
}

// NewView builds the help [View] for cmd. cl supplies the command's
// positional arguments.
func NewView(cmd *cobra.Command, cl *arg.CommandLine) View {
	if cl == nil {
		cl = (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	}
	return View{
		Name:          cmd.CommandPath(),
		Description:   descriptionOf(cmd),
		Usage:         usageLineOf(cmd),
		Examples:      examplesOf(cmd),
		CommandGroups: commandGroupsOf(cmd),
		Arguments:     argumentsOf(cl),
		FlagGroups:    flagGroupsOf(cl),
		Hint:          hintOf(cmd),
	}
}

// descriptionOf returns the long description of cmd, falling back to the short
// summary when no long description is set.
func descriptionOf(cmd *cobra.Command) string {
	if cmd.Long != "" {
		return cmd.Long
	}
	return cmd.Short
}

// usageLineOf returns the usage line for cmd, qualified by the path of the
// command it is reached through.
func usageLineOf(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return cmd.Parent().CommandPath() + " " + cmd.Use
	}
	return cmd.Use
}

// examplesOf returns the non-blank example lines of cmd, in order.
func examplesOf(cmd *cobra.Command) []string {
	if cmd.Example == "" {
		return nil
	}
	var examples []string
	for line := range strings.SplitSeq(cmd.Example, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		examples = append(examples, line)
	}
	return examples
}

// commandGroupsOf returns the visible subcommands of cmd organised by their
// cobra group, with ungrouped subcommands collected under [additionalCommands]
// last. Groups with no visible subcommands are omitted.
func commandGroupsOf(cmd *cobra.Command) []CommandGroup {
	var groups []CommandGroup
	for _, group := range cmd.Groups() {
		commands := commandsInGroup(cmd, group.ID)
		if len(commands) == 0 {
			continue
		}
		groups = append(groups, CommandGroup{Title: group.Title, Commands: commands})
	}
	if commands := commandsInGroup(cmd, ""); len(commands) > 0 {
		groups = append(groups, CommandGroup{Title: additionalCommands, Commands: commands})
	}
	return groups
}

// commandsInGroup returns the visible subcommands of parent whose group id
// matches id, in cobra's listing order.
func commandsInGroup(parent *cobra.Command, id string) []Command {
	var commands []Command
	for _, c := range parent.Commands() {
		if !visibleCommand(c) || c.GroupID != id {
			continue
		}
		commands = append(commands, Command{Name: c.Name(), Summary: c.Short})
	}
	return commands
}

// visibleCommand reports whether c should appear in a help listing.
func visibleCommand(c *cobra.Command) bool {
	return c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand()
}

// argumentsOf returns the arguments registered on cl: the positionals in
// registration order, followed by the unmatched-argument binding when one is
// registered. The unmatched entry is variadic, since it claims every remaining
// argument rather than a single slot.
func argumentsOf(cl *arg.CommandLine) []ArgumentInfo {
	var arguments []ArgumentInfo
	for _, p := range argdef.Positionals((*argdef.CommandLine)(cl)) {
		arguments = append(arguments, ArgumentInfo{
			Name:     p.Name,
			Type:     p.Type,
			Usage:    p.Usage,
			Required: p.Required,
		})
	}
	if u := argdef.GetUnmatched((*argdef.CommandLine)(cl)); u != nil {
		arguments = append(arguments, ArgumentInfo{
			Name:     u.Name,
			Type:     u.Type,
			Usage:    u.Usage,
			Required: u.Required,
			Variadic: true,
		})
	}
	return arguments
}

// flagGroupsOf returns the visible flag groups of cl. Fully hidden groups
// are omitted, and hidden flags within a group are excluded.
func flagGroupsOf(cl *arg.CommandLine) []FlagGroup {
	var groups []FlagGroup
	for _, group := range cl.Groups() {
		if group.Hidden() {
			continue
		}
		// A non-hidden group always has at least one non-hidden flag, since
		// [arg.FlagGroup.Hidden] reports true only when every flag is hidden.
		var flags []FlagInfo
		for _, f := range group.Flags {
			if f.Hidden() {
				continue
			}
			flags = append(flags, flagInfoOf(f))
		}
		groups = append(groups, FlagGroup{Title: group.Name, Flags: flags})
	}
	return groups
}

// flagInfoOf extracts the display information for a single flag.
func flagInfoOf(f *arg.FlagArg) FlagInfo {
	return FlagInfo{
		Shorthand: f.Shorthand(),
		Name:      f.Name(),
		Type:      flagTypeOf(f),
		Usage:     f.Usage(),
	}
}

// flagTypeOf returns the type name to display for f, or an empty string for
// boolean flags, which take no value argument.
func flagTypeOf(f *arg.FlagArg) string {
	if t := f.Type(); t != "bool" {
		return t
	}
	return ""
}

// hintOf returns the trailing help advice for cmd, shown only when cmd has
// visible subcommands.
func hintOf(cmd *cobra.Command) Hint {
	if len(commandGroupsOf(cmd)) == 0 {
		return Hint{}
	}
	return Hint{Show: true, Path: cmd.CommandPath()}
}
