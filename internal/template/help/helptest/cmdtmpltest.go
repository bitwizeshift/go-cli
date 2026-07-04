package helptest

import (
	"fmt"
	"time"

	"github.com/bitwizeshift/go-cli/flag"
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
		cases = append(cases, Case{
			Name:    fmt.Sprintf("sub_%d.golden.txt", width),
			Columns: width,
			Command: Subcommand(),
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
	root.AddCommand(
		initCommand(),
		addCommand(),
		syncCommand(),
		completionCommand(),
	)
	return root
}

// Subcommand returns the flag-rich "sync" subcommand of a freshly built [Root],
// with its parent linkage intact.
func Subcommand() *cobra.Command {
	for _, c := range Root().Commands() {
		if c.Name() == subcommandName {
			return c
		}
	}
	return nil
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

func syncCommand() *cobra.Command {
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
	registerFlags(cmd)
	return cmd
}

// registerFlags registers the sync command's flags across two named groups.
func registerFlags(cmd *cobra.Command) {
	fs := cmd.Flags()
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

	flag.AddToGroup("Connection Flags",
		flag.Add(fs, "auth-token", &authToken,
			flag.Shorthand("T"),
			flag.Usage("auth token used to authenticate with the remote"),
		),
		flag.Add(fs, "force", &force,
			flag.Shorthand("f"),
			flag.Usage("overwrite any item already present in the vault"),
		),
		flag.Add(fs, "parallel", &parallel,
			flag.Shorthand("p"),
			flag.Usage("number of transfers to run at once"),
		),
		flag.Add(fs, "remote", &remote,
			flag.Shorthand("r"),
			flag.Usage("base URL of the remote to synchronize with"),
		),
		flag.Add(fs, "timeout", &timeout,
			flag.Shorthand("t"),
			flag.Type("duration"),
			flag.Usage("maximum time to wait for the sync to finish"),
		),
	)
	flag.AddToGroup("Output Flags",
		flag.Add(fs, "exclude-dependencies", &excludeDeps,
			flag.Usage("skip synchronizing the transitive dependencies of the item"),
		),
		flag.Add(fs, "state-dir", &stateDir,
			flag.Usage("directory in which sync state is stored"),
		),
		flag.Add(fs, "log", &logFile,
			flag.Usage("file to write sync progress logs to"),
		),
		flag.Add(fs, "no-progress", &noProgress,
			flag.Usage("disable the interactive progress bar"),
		),
		flag.Add(fs, "verbose", &verbose,
			flag.Shorthand("v"),
			flag.Usage("print additional diagnostic output while running"),
		),
	)
}

func noop(*cobra.Command, []string) {}
