package argtest_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

// errDecode is reported by failDecode to observe a binding failure in [Parse].
var errDecode = errors.New("decode failed")

// failDecode is a string decoder that always fails with errDecode.
func failDecode([]byte) (string, error) { return "", errDecode }

type FakeT struct {
	testing.TB
	IsHelper bool
	Fatals   []string
}

func (ft *FakeT) Helper() {
	ft.IsHelper = true
}

func (ft *FakeT) Fatalf(format string, args ...any) {
	ft.Fatals = append(ft.Fatals, fmt.Sprintf(format, args...))
}

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		wantFatals int
	}{
		{
			name:       "ValidArgsDoNotFail",
			args:       []string{"--flag"},
			wantFatals: 0,
		},
		{
			name:       "InvalidArgsFail",
			args:       []string{"--unknown"},
			wantFatals: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			cl.Add(arg.Flag("flag", new(bool)))
			ft := &FakeT{}

			// Act
			argtest.Parse(ft, cl, tc.args...)

			// Assert
			if got, want := len(ft.Fatals), tc.wantFatals; !cmp.Equal(got, want) {
				t.Errorf("Parse(...) fatals = %d, want %d", got, want)
			}
		})
	}
}

func TestAllFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	verbose := arg.Flag("verbose", new(bool), arg.Shorthand("v"))
	name := arg.Flag("name", new(string))
	region := arg.Flag("region", new(string))
	novalue := arg.Flag("novalue", new(int))
	cl.Add(verbose, name, region, novalue)
	arg.MarkRequired(name)
	arg.MarkRequiredTogether(verbose, name)
	arg.MarkMutuallyExclusive(verbose, novalue)
	arg.MarkOneRequired(name, novalue)
	arg.Group("Location Flags", novalue)

	// Act
	flags := argtest.AllFlags(cl)

	// Assert
	want := []*argtest.Flag{
		{
			Long:            "name",
			Type:            "string",
			Required:        true,
			RequiredWith:    []string{"name", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long:            "novalue",
			Type:            "int",
			Group:           "Location Flags",
			ExclusiveWith:   []string{"novalue", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long: "region",
			Type: "string",
		},
		{
			Long:          "verbose",
			Short:         "v",
			Type:          "bool",
			RequiredWith:  []string{"name", "verbose"},
			ExclusiveWith: []string{"novalue", "verbose"},
		},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestAllPositionals(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var src, dst string
	cl.Add(
		arg.Positional("src", 0, &src, arg.Usage("source path")),
		arg.Positional("dst", 1, &dst, arg.Type("path"), arg.Usage("destination path")),
	)

	// Act
	positionals := argtest.AllPositionals(cl)

	// Assert
	want := []*argtest.Positional{
		{
			Index: 0,
			Name:  "src",
			Type:  "string",
			Usage: "source path",
		},
		{
			Index: 1,
			Name:  "dst",
			Type:  "path",
			Usage: "destination path",
		},
	}
	if got, want := positionals, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllPositionals(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestGetUnmatched(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		want    *argtest.Unmatched
	}{
		{
			name:    "ReportsElementTypeAndUsage",
			options: []arg.Option{arg.Usage("remaining paths")},
			want: &argtest.Unmatched{
				Type:  "string",
				Usage: "remaining paths",
			},
		}, {
			name:    "ReportsOverriddenType",
			options: []arg.Option{arg.Type("path"), arg.Usage("remaining paths")},
			want: &argtest.Unmatched{
				Type:  "path",
				Usage: "remaining paths",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var rest []string
			cl.Add(arg.Unmatched(&rest, tc.options...))

			// Act
			unmatched := argtest.GetUnmatched(cl)

			// Assert
			if got, want := unmatched, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GetUnmatched(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
			}
		})
	}
}

func TestGetUnmatched_Unbound_ReportsNone(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()

	// Act
	unmatched := argtest.GetUnmatched(cl)

	// Assert
	if got, want := unmatched, (*argtest.Unmatched)(nil); !cmp.Equal(got, want) {
		t.Errorf("GetUnmatched(...) = %v, want %v", got, want)
	}
}

func TestParse_BindsPositionals(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var name string
	var rest []string
	cl.Add(
		arg.Positional("name", 0, &name),
		arg.Unmatched(&rest),
	)
	ft := &FakeT{}

	// Act
	argtest.Parse(ft, cl, "alpha", "beta", "gamma")

	// Assert
	if got, want := len(ft.Fatals), 0; !cmp.Equal(got, want) {
		t.Fatalf("Parse(...) fatals = %d, want %d", got, want)
	}
	if got, want := name, "alpha"; !cmp.Equal(got, want) {
		t.Errorf("positional name = %q, want %q", got, want)
	}
	if got, want := rest, []string{"beta", "gamma"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("unmatched arguments = %v, want %v", got, want)
	}
}

func TestParse_BindError(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	cl.Add(arg.Positional("value", 0, new(string), arg.UnmarshalWith(failDecode)))
	ft := &FakeT{}

	// Act
	argtest.Parse(ft, cl, "x")

	// Assert
	if got, want := len(ft.Fatals), 1; !cmp.Equal(got, want) {
		t.Errorf("Parse(...) fatals = %d, want %d", got, want)
	}
}

func TestLongFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	var b bool
	cl := argtest.NewCommandLine()
	cl.Add(
		arg.Flag("alpha", &b, arg.Shorthand("a")),
		arg.Flag("beta", &b),
	)

	// Act
	long := argtest.LongFlags(cl)

	// Assert
	if got, want := long, []string{"alpha", "beta"}; !cmp.Equal(got, want) {
		t.Errorf("LongFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestShortFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	cl.Add(
		arg.Flag("alpha", new(bool), arg.Shorthand("a")),
		arg.Flag("beta", new(bool)),
	)

	// Act
	short := argtest.ShortFlags(cl)

	// Assert
	if got, want := short, []string{"a", ""}; !cmp.Equal(got, want) {
		t.Errorf("ShortFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
