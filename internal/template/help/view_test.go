package help_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/template/help"
)

func TestNewView(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		command *cobra.Command
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			command := tc.command

			// Act
			view := help.NewView(command)

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
	fs := cmd.Flags()
	var force bool
	var zulu, yankee, beta, gamma string
	forceFlag := flag.Add(fs, "force", &force, flag.Shorthand("f"), flag.Usage("force it"))
	zuluFlag := flag.Add(fs, "zulu", &zulu, flag.Usage("z option"))
	betaFlag := flag.Add(fs, "beta", &beta, flag.Usage("b option"))
	flag.Add(fs, "yankee", &yankee, flag.Usage("y option"))
	gammaFlag := flag.Add(fs, "gamma", &gamma, flag.Usage("g option"))
	annotation.AddToGroup("Zeta Flags", forceFlag, zuluFlag, betaFlag)
	annotation.AddToGroup("Secret Flags", gammaFlag)
	betaFlag.Hidden = true
	gammaFlag.Hidden = true
	return cmd
}

func noop(*cobra.Command, []string) {}
