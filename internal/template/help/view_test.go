package help_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/template/help"
)

func TestNewView(t *testing.T) {
	t.Parallel()

	positionalCmd, positionalCL := commandWithPositionals()
	unmatchedCmd, unmatchedCL := commandWithUnmatched()
	unmatchedOnlyCmd, unmatchedOnlyCL := commandWithUnmatchedOnly()

	testCases := []struct {
		name    string
		command *cobra.Command
		cl      *arg.CommandLine
		want    help.View
	}{
		{
			name:    "ShortUsedWhenNoLongDescription",
			command: leaf("tiny", "", "a tiny command"),
			want: help.View{
				Name:        "tiny",
				Description: "a tiny command",
				Usage:       "tiny [flags]",
			},
		}, {
			name:    "LongPreferredOverShort",
			command: leaf("tiny", "the long description", "short summary"),
			want: help.View{
				Name:        "tiny",
				Description: "the long description",
				Usage:       "tiny [flags]",
			},
		}, {
			name:    "SubcommandUsagePrefixedWithParentPath",
			command: subcommand(),
			want: help.View{
				Name:        "app sub",
				Description: "the sub",
				Usage:       "app sub <arg> [flags]",
			},
		}, {
			name:    "ExamplesSplitAndBlankLinesDropped",
			command: commandWithExamples(),
			want: help.View{
				Name:        "note",
				Description: "note",
				Usage:       "note [flags]",
				Examples:    []string{"note first", "note second"},
			},
		}, {
			name:    "CommandGroupsWithHiddenEmptyAndAdditional",
			command: commandWithGroups(),
			want: help.View{
				Name:        "app",
				Description: "app",
				Usage:       "app <command> [flags]",
				CommandGroups: []help.CommandGroup{
					{
						Title:    "Tools",
						Commands: []help.Command{{Name: "build", Summary: "build it"}},
					}, {
						Title:    "Additional Commands",
						Commands: []help.Command{{Name: "tidy", Summary: "tidy up"}},
					},
				},
				Hint: help.Hint{Show: true, Path: "app"},
			},
		}, {
			name:    "FlagGroupsWithHiddenFilteringAndGeneralLast",
			command: commandWithFlagGroups(),
			want: help.View{
				Name:        "svc",
				Description: "svc",
				Usage:       "svc [flags]",
				FlagGroups: []help.FlagGroup{
					{
						Title: "Zeta Flags",
						Flags: []help.FlagInfo{
							{Shorthand: "f", Name: "force", Type: "", Usage: "force it"},
							{Shorthand: "", Name: "zulu", Type: "string", Usage: "z option"},
						},
					}, {
						Title: "General Flags",
						Flags: []help.FlagInfo{
							{Shorthand: "", Name: "yankee", Type: "string", Usage: "y option"},
						},
					},
				},
			},
		}, {
			name:    "PositionalsListedInRegistrationOrder",
			command: positionalCmd,
			cl:      positionalCL,
			want: help.View{
				Name:        "cp",
				Description: "cp",
				Usage:       "cp <src> <dst> [flags]",
				Arguments: []help.ArgumentInfo{
					{Name: "src", Type: "string", Usage: "source path", Required: true},
					{Name: "dst", Type: "string", Usage: "destination path"},
				},
			},
		}, {
			name:    "UnmatchedListedAfterPositionals",
			command: unmatchedCmd,
			cl:      unmatchedCL,
			want: help.View{
				Name:        "rm",
				Description: "rm",
				Usage:       "rm <src> [flags]",
				Arguments: []help.ArgumentInfo{
					{Name: "src", Type: "string", Usage: "source path"},
					{Type: "string...", Usage: "additional paths"},
				},
			},
		}, {
			name:    "UnmatchedAloneReportsElementType",
			command: unmatchedOnlyCmd,
			cl:      unmatchedOnlyCL,
			want: help.View{
				Name:        "sum",
				Description: "sum",
				Usage:       "sum [flags]",
				Arguments: []help.ArgumentInfo{
					{Type: "int...", Usage: "values to sum", Required: true},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			command := tc.command

			// Act
			view := help.NewView(command, tc.cl)

			// Assert
			if got, want := view, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("NewView(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func leaf(use, long, short string) *cobra.Command {
	return &cobra.Command{Use: use, Long: long, Short: short, Run: noop}
}

func subcommand() *cobra.Command {
	root := &cobra.Command{Use: "app", Short: "app"}
	sub := &cobra.Command{Use: "sub <arg>", Short: "the sub", Run: noop}
	root.AddCommand(sub)
	return sub
}

func commandWithExamples() *cobra.Command {
	return &cobra.Command{
		Use:   "note",
		Short: "note",
		Run:   noop,
		Example: `note first

note second`,
	}
}

func commandWithGroups() *cobra.Command {
	root := &cobra.Command{Use: "app <command>", Short: "app"}
	root.AddGroup(&cobra.Group{ID: "tools", Title: "Tools"})
	root.AddGroup(&cobra.Group{ID: "empty", Title: "Empty"})
	root.AddCommand(&cobra.Command{Use: "build", Short: "build it", GroupID: "tools", Run: noop})
	root.AddCommand(&cobra.Command{Use: "secret", Short: "s", GroupID: "tools", Hidden: true, Run: noop})
	root.AddCommand(&cobra.Command{Use: "tidy", Short: "tidy up", Run: noop})
	return root
}

func commandWithFlagGroups() *cobra.Command {
	cmd := &cobra.Command{Use: "svc", Short: "svc", Run: noop}
	cl := (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	var force bool
	var zulu, yankee, beta, gamma string
	forceFlag := arg.Flag("force", &force, arg.Shorthand("f"), arg.Usage("force it"))
	zuluFlag := arg.Flag("zulu", &zulu, arg.Usage("z option"))
	betaFlag := arg.Flag("beta", &beta, arg.Usage("b option"), arg.Hidden())
	yankeeFlag := arg.Flag("yankee", &yankee, arg.Usage("y option"))
	gammaFlag := arg.Flag("gamma", &gamma, arg.Usage("g option"), arg.Hidden())
	cl.Add(forceFlag, zuluFlag, betaFlag, yankeeFlag, gammaFlag)
	arg.Group("Zeta Flags", forceFlag, zuluFlag, betaFlag)
	arg.Group("Secret Flags", gammaFlag)
	return cmd
}

func commandWithPositionals() (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{Use: "cp <src> <dst>", Short: "cp", Run: noop}
	cl := (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	var src, dst string
	cl.Add(
		arg.Positional("src", 0, &src, arg.Usage("source path"), arg.Required()),
		arg.Positional("dst", 1, &dst, arg.Usage("destination path")),
	)
	return cmd, cl
}

func commandWithUnmatched() (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{Use: "rm <src>", Short: "rm", Run: noop}
	cl := (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	var src string
	var rest []string
	cl.Add(
		arg.Positional("src", 0, &src, arg.Usage("source path")),
		arg.Unmatched(&rest, arg.Usage("additional paths")),
	)
	return cmd, cl
}

func commandWithUnmatchedOnly() (*cobra.Command, *arg.CommandLine) {
	cmd := &cobra.Command{Use: "sum", Short: "sum", Run: noop}
	cl := (*arg.CommandLine)(argdef.FromFlagSet(cmd.Flags()))
	var values []int
	cl.Add(arg.Unmatched(&values, arg.Usage("values to sum"), arg.Required()))
	return cmd, cl
}

func noop(*cobra.Command, []string) {}
