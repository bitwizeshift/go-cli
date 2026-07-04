package flagtest_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag/flagtest"
	"github.com/bitwizeshift/go-cli/internal/annotation"
)

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
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			fs.Bool("flag", false, "")
			ft := &FakeT{}

			// Act
			flagtest.Parse(ft, fs, tc.args...)

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
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.BoolP("verbose", "v", false, "")
	fs.String("name", "", "")
	fs.String("region", "", "")
	fs.AddFlag(&pflag.Flag{Name: "novalue"})
	annotation.MarkRequired(fs.Lookup("name"))
	annotation.MarkRequiredTogether(fs.Lookup("verbose"), fs.Lookup("name"))
	annotation.MarkMutuallyExclusive(fs.Lookup("verbose"), fs.Lookup("novalue"))
	annotation.MarkOneRequired(fs.Lookup("name"), fs.Lookup("novalue"))
	annotation.AddToGroup("Location Flags", fs.Lookup("region"))

	// Act
	flags := flagtest.AllFlags(fs)

	// Assert
	want := []*flagtest.Flag{
		{
			Long:            "name",
			Type:            "string",
			Required:        true,
			RequiredWith:    []string{"name", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long:            "novalue",
			ExclusiveWith:   []string{"novalue", "verbose"},
			OneRequiredWith: []string{"name", "novalue"},
		},
		{
			Long:  "region",
			Type:  "string",
			Group: "Location Flags",
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

func TestLongFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.BoolP("alpha", "a", false, "")
	fs.Bool("beta", false, "")

	// Act
	long := flagtest.LongFlags(fs)

	// Assert
	if got, want := long, []string{"alpha", "beta"}; !cmp.Equal(got, want) {
		t.Errorf("LongFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestShortFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.BoolP("alpha", "a", false, "")
	fs.Bool("beta", false, "")

	// Act
	short := flagtest.ShortFlags(fs)

	// Assert
	if got, want := short, []string{"a", ""}; !cmp.Equal(got, want) {
		t.Errorf("ShortFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
