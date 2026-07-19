package argdef_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/internal/argdef"
)

// addStringFlags registers a string flag on fs for each name, and returns the
// registered flags.
func addStringFlags(fs *pflag.FlagSet, names []string) []*pflag.Flag {
	flags := make([]*pflag.Flag, 0, len(names))
	for _, name := range names {
		fs.String(name, "", "")
		flags = append(flags, fs.Lookup(name))
	}
	return flags
}

// addBoolFlags registers a boolean flag on fs for each name, and returns the
// registered flags.
func addBoolFlags(fs *pflag.FlagSet, names []string) []*pflag.Flag {
	flags := make([]*pflag.Flag, 0, len(names))
	for _, name := range names {
		fs.Bool(name, false, "")
		flags = append(flags, fs.Lookup(name))
	}
	return flags
}

// hideFlags marks every flag in flags as hidden, and returns them.
func hideFlags(flags []*pflag.Flag) []*pflag.Flag {
	for _, f := range flags {
		f.Hidden = true
	}
	return flags
}

func TestUsage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		optionalFlags    []string
		requiredFlags    []string
		requiredBools    []string
		hiddenFlags      []string
		hiddenRequired   []string
		oneRequired      []string
		hiddenOneOf      []string
		requiredTogether []string
		positionals      []*argdef.Positional
		unmatched        *argdef.Unmatched
		operands         []string
		want             string
	}{
		{
			name: "NoArguments",
			want: "",
		}, {
			name:          "RequiredFlagSpelledOut",
			requiredFlags: []string{"token"},
			want:          "--token <string>",
		}, {
			name:          "RequiredBoolFlagTakesNoValue",
			requiredBools: []string{"force"},
			want:          "--force",
		}, {
			name:          "RequiredFlagsSortedByName",
			requiredFlags: []string{"zulu", "alpha"},
			want:          "--alpha <string> --zulu <string>",
		}, {
			name:          "OptionalFlagReportedAsPlaceholder",
			optionalFlags: []string{"verbose"},
			want:          "[flags]",
		}, {
			name:          "RequiredFlagPrecedesPlaceholder",
			requiredFlags: []string{"token"},
			optionalFlags: []string{"verbose"},
			want:          "--token <string> [flags]",
		}, {
			name:        "HiddenOptionalFlagOmitsPlaceholder",
			hiddenFlags: []string{"debug"},
			want:        "",
		}, {
			name:           "HiddenRequiredFlagOmitted",
			hiddenRequired: []string{"internal"},
			want:           "",
		}, {
			name:        "OneRequiredGroupRendersAlternation",
			oneRequired: []string{"url", "remote"},
			want:        "(--remote <string> | --url <string>)",
		}, {
			name:          "OneRequiredGroupPrecedesPlaceholder",
			oneRequired:   []string{"url", "remote"},
			optionalFlags: []string{"verbose"},
			want:          "(--remote <string> | --url <string>) [flags]",
		}, {
			name:        "OneRequiredGroupWithSingleVisibleMember",
			oneRequired: []string{"remote"},
			hiddenOneOf: []string{"url"},
			want:        "--remote <string>",
		}, {
			name:        "OneRequiredGroupFullyHiddenOmitted",
			hiddenOneOf: []string{"url", "remote"},
			want:        "",
		}, {
			name:             "RequiredTogetherFoldsIntoPlaceholder",
			requiredTogether: []string{"user", "password"},
			want:             "[flags]",
		}, {
			name: "RequiredPositional",
			positionals: []*argdef.Positional{
				{Name: "src", Index: 0, Required: true},
			},
			want: "<src>",
		}, {
			name: "OptionalPositional",
			positionals: []*argdef.Positional{
				{Name: "dst", Index: 0},
			},
			want: "[dst]",
		}, {
			name: "PositionalsInRegistrationOrder",
			positionals: []*argdef.Positional{
				{Name: "src", Index: 0, Required: true},
				{Name: "dst", Index: 1},
			},
			want: "<src> [dst]",
		}, {
			name:      "RequiredUnmatchedIsVariadic",
			unmatched: &argdef.Unmatched{Name: "names", Type: "string", Required: true},
			want:      "<names>...",
		}, {
			name:      "OptionalUnmatchedIsVariadic",
			unmatched: &argdef.Unmatched{Name: "names", Type: "string"},
			want:      "[names...]",
		}, {
			name:     "OperandRenderedAlone",
			operands: []string{"<command>"},
			want:     "<command>",
		}, {
			name:          "OperandFollowsFlagsAndPrecedesPositionals",
			requiredFlags: []string{"token"},
			operands:      []string{"<command>"},
			positionals: []*argdef.Positional{
				{Name: "src", Index: 0, Required: true},
			},
			want: "--token <string> <command> <src>",
		}, {
			name:          "EveryArgumentKind",
			requiredFlags: []string{"token"},
			optionalFlags: []string{"verbose"},
			positionals: []*argdef.Positional{
				{Name: "src", Index: 0, Required: true},
				{Name: "dst", Index: 1},
			},
			unmatched: &argdef.Unmatched{Name: "names", Type: "string"},
			want:      "--token <string> <src> [dst] [names...] [flags]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argdef.New()
			fs := argdef.Flags(cl)
			addStringFlags(fs, tc.optionalFlags)
			argdef.MarkRequired(addStringFlags(fs, tc.requiredFlags)...)
			argdef.MarkRequired(addBoolFlags(fs, tc.requiredBools)...)
			hideFlags(addStringFlags(fs, tc.hiddenFlags))
			argdef.MarkRequired(hideFlags(addStringFlags(fs, tc.hiddenRequired))...)
			argdef.MarkRequiredTogether(addStringFlags(fs, tc.requiredTogether)...)
			oneOf := addStringFlags(fs, tc.oneRequired)
			oneOf = append(oneOf, hideFlags(addStringFlags(fs, tc.hiddenOneOf))...)
			argdef.MarkOneRequired(oneOf...)
			for _, p := range tc.positionals {
				argdef.AddPositional(cl, p)
			}
			if tc.unmatched != nil {
				argdef.SetUnmatched(cl, tc.unmatched)
			}

			// Act
			usage := argdef.Usage(cl, tc.operands...)

			// Assert
			if got, want := usage, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Usage(...) = %q, want %q", got, want)
			}
		})
	}
}
