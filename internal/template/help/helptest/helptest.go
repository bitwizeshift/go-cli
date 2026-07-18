package helptest

import (
	"fmt"
	"time"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argreg"
	"github.com/spf13/cobra"
)

// Case is one golden-file scenario: a command rendered at a column width, keyed
// by the golden file that records its output.
type Case struct {
	// Name is the golden file name that stores the rendered output.
	Name string

	// Columns is the width the command is rendered at.
	Columns int

	// Command is the command to render.
	Command *cobra.Command

	// CL is the argument registry supplying Command's positional arguments. It is
	// nil for a command with none.
	CL *arg.CommandLine
}

// Cases returns the golden scenarios shared by the generator and the golden
// test: the root command and the flag-rich subcommand, each at 60, 80, and 100
// columns.
func Cases() []Case {
	widths := []int{60, 80, 100}
	cases := make([]Case, 0, 2*len(widths))
	for _, width := range widths {
		cases = append(cases, Case{
			Name:    fmt.Sprintf("root_%d.golden.txt", width),
			Columns: width,
			Command: Root(),
		})
	}
	for _, width := range widths {
		sub, cl := Subcommand()
		cases = append(cases, Case{
			Name:    fmt.Sprintf("sub_%d.golden.txt", width),
			Columns: width,
			Command: sub,
			CL:      cl,
		})
	}
	return cases
}

const (
	itemGroup   = "item"
	remoteGroup = "remote"

	// subcommandName is the name of the flag-rich subcommand returned by
	// [Subcommand].
	subcommandName = "sync"
)

// Root returns the root command of the fixture hierarchy.
func Root() *cobra.Command {
	root, _ := buildRoot()
	return root
}

// Subcommand returns the flag-rich "sync" subcommand of a freshly built root,
// with its parent linkage intact, alongside the argument registry supplying its
// positional arguments.
func Subcommand() (*cobra.Command, *arg.CommandLine) {
	root, cl := buildRoot()
	for _, c := range root.Commands() {
		if c.Name() == subcommandName {
			return c, cl
		}
	}
	return nil, nil
}

// buildRoot assembles the fixture hierarchy, returning the root command and the
// argument registry of its flag-rich "sync" subcommand.
func buildRoot() (*cobra.Command, *arg.CommandLine) {
	root := &cobra.Command{
		Use:   "example-cli <command>",
		Short: "example-cli is a small CLI that demonstrates the help renderer",
		Long: "example-cli is a lightweight command-line application used to " +
			"demonstrate the custom help renderer, including coloured sections, " +
			"wrapped prose, and flags and commands organised into named groups.",
		Example: `example-cli init
example-cli sync origin main --remote https://vault.example.org`,
	}
	root.AddGroup(
		&cobra.Group{ID: itemGroup, Title: "Item Commands"},
		&cobra.Group{ID: remoteGroup, Title: "Remote Commands"},
	)
	sync, cl := syncCommand()
	root.AddCommand(
		initCommand(),
		addCommand(),
		sync,
		completionCommand(),
	)
	return root, cl
}

func initCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "init",
		GroupID: itemGroup,
		Short:   "Initialize a new vault in the current directory",
		Run:     noop,
	}
}

func addCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "add <name>",
		GroupID: itemGroup,
		Short:   "Add an item to the vault and write the updated contents to disk",
		Run:     noop,
	}
}

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate the autocompletion script for the specified shell",
		Run:   noop,
	}
}

func syncCommand() (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{
		Use:     subcommandName + " <remote> <ref>",
		GroupID: remoteGroup,
		Short:   "Synchronize the vault with a remote",
		Long: "Synchronize the vault with a remote registry, resolving and " +
			"caching each referenced item and, unless told otherwise, its " +
			"transitive dependencies.",
		Example: `example-cli sync origin main
example-cli sync origin v1.2.0 --force --timeout 1m`,
		Run: noop,
	}
	return cmd, registerArgs(cmd)
}

// registerArgs registers the sync command's positional arguments and its flags,
// the latter across two named groups. It returns the argument registry so the
// positional arguments can be rendered in help.
func registerArgs(cmd *cobra.Command) *arg.CommandLine {
	cl := (*arg.CommandLine)(argreg.FromFlagSet(cmd.Flags()))
	var (
		remoteRef string
		ref       string
	)
	arg.Positional(cl, "remote", 0, &remoteRef,
		arg.Usage("name of the remote to synchronize with"),
	)
	arg.Positional(cl, "ref", 1, &ref,
		arg.Usage("reference within the remote to synchronize"),
	)
	var (
		authToken string
		force     bool
		parallel  int
		remote    string
		timeout   time.Duration

		excludeDeps bool
		stateDir    string
		logFile     string
		noProgress  bool
		verbose     bool
	)

	arg.AddToGroup("Connection Flags",
		arg.AddFlag(cl, "auth-token", &authToken,
			arg.Shorthand("T"),
			arg.Usage("auth token used to authenticate with the remote"),
		),
		arg.AddFlag(cl, "force", &force,
			arg.Shorthand("f"),
			arg.Usage("overwrite any item already present in the vault"),
		),
		arg.AddFlag(cl, "parallel", &parallel,
			arg.Shorthand("p"),
			arg.Usage("number of transfers to run at once"),
		),
		arg.AddFlag(cl, "remote", &remote,
			arg.Shorthand("r"),
			arg.Usage("base URL of the remote to synchronize with"),
		),
		arg.AddFlag(cl, "timeout", &timeout,
			arg.Shorthand("t"),
			arg.Type("duration"),
			arg.Usage("maximum time to wait for the sync to finish"),
		),
	)
	arg.AddToGroup("Output Flags",
		arg.AddFlag(cl, "exclude-dependencies", &excludeDeps,
			arg.Usage("skip synchronizing the transitive dependencies of the item"),
		),
		arg.AddFlag(cl, "state-dir", &stateDir,
			arg.Usage("directory in which sync state is stored"),
		),
		arg.AddFlag(cl, "log", &logFile,
			arg.Usage("file to write sync progress logs to"),
		),
		arg.AddFlag(cl, "no-progress", &noProgress,
			arg.Usage("disable the interactive progress bar"),
		),
		arg.AddFlag(cl, "verbose", &verbose,
			arg.Shorthand("v"),
			arg.Usage("print additional diagnostic output while running"),
		),
	)
	return cl
}

func noop(*cobra.Command, []string) {}
