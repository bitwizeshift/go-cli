package flagtest_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
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
			registry := flagtest.NewRegistry()
			flag.Add(registry, "flag", new(bool))
			ft := &FakeT{}

			// Act
			flagtest.Parse(ft, registry, tc.args...)

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
	registry := flagtest.NewRegistry()
	verbose := flag.Add(registry, "verbose", new(bool), flag.Shorthand("v"))
	name := flag.Add(registry, "name", new(string))
	flag.Add(registry, "region", new(string))
	novalue := flag.Add(registry, "novalue", new(int))
	flag.MarkRequired(name)
	flag.MarkRequiredTogether(verbose, name)
	flag.MarkMutuallyExclusive(verbose, novalue)
	flag.MarkOneRequired(name, novalue)
	flag.AddToGroup("Location Flags", novalue)

	// Act
	flags := flagtest.AllFlags(registry)

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

func TestLongFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	var b bool
	registry := flagtest.NewRegistry()
	flag.Add(registry, "alpha", &b, flag.Shorthand("a"))
	flag.Add(registry, "beta", &b)

	// Act
	long := flagtest.LongFlags(registry)

	// Assert
	if got, want := long, []string{"alpha", "beta"}; !cmp.Equal(got, want) {
		t.Errorf("LongFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestShortFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	flag.Add(registry, "alpha", new(bool), flag.Shorthand("a"))
	flag.Add(registry, "beta", new(bool))

	// Act
	short := flagtest.ShortFlags(registry)

	// Assert
	if got, want := short, []string{"a", ""}; !cmp.Equal(got, want) {
		t.Errorf("ShortFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
