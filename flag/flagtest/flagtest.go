package flagtest

import (
	"slices"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/flagreg"
)

// Parse parses args into the flags registered in registry, failing the test via
// [testing.TB.Fatalf] if parsing returns an error.
func Parse(t testing.TB, registry *flag.Registry, args ...string) {
	t.Helper()

	if err := registry.FlagSet().Parse(args); err != nil {
		t.Fatalf("Parse(...): unexpected error: %v", err)
	}
}

// Flag is a small wrapper around the values assigned to a registered flag, for
// property-testing purposes.
type Flag struct {
	Long  string
	Short string
	Type  string
	Group string

	ExclusiveWith   []string
	RequiredWith    []string
	OneRequiredWith []string
	Required        bool
}

// NewRegistry returns a new [flag.Registry] that can be used for testing.
func NewRegistry() *flag.Registry {
	reg := flagreg.New()
	return (*flag.Registry)(reg)
}

// AllFlags returns a [Flag] for every flag registered in registry, sorted by
// long name.
func AllFlags(registry *flag.Registry) []*Flag {
	var result []*Flag
	for _, f := range registry.Flags() {
		fv := Flag{
			Long:  f.Name(),
			Short: f.Shorthand(),
			Type:  f.Type(),
			Group: f.Group(),

			ExclusiveWith:   f.MutuallyExclusiveWith(),
			RequiredWith:    f.RequiredWith(),
			OneRequiredWith: f.OneRequiredWith(),
			Required:        f.Required(),
		}
		result = append(result, &fv)
	}
	slices.SortFunc(result, func(lhs, rhs *Flag) int {
		return strings.Compare(lhs.Long, rhs.Long)
	})
	return result
}

// LongFlags returns the long names of every flag registered in fs, sorted.
func LongFlags(registry *flag.Registry) []string {
	all := AllFlags(registry)
	long := make([]string, 0, len(all))
	for _, f := range all {
		long = append(long, f.Long)
	}
	return long
}

// ShortFlags returns the shorthand names of every flag registered in fs, in
// long-name order.
func ShortFlags(registry *flag.Registry) []string {
	all := AllFlags(registry)
	short := make([]string, 0, len(all))
	for _, f := range all {
		short = append(short, f.Short)
	}
	return short
}
