package flagtest

import (
	"slices"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/spf13/pflag"
)

// Parse parses args into fs, failing the test via [testing.TB.Fatalf] if
// parsing returns an error.
func Parse(t testing.TB, fs *pflag.FlagSet, args ...string) {
	t.Helper()

	if err := fs.Parse(args); err != nil {
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

// AllFlags returns a [Flag] for every flag registered in fs, sorted by long
// name. Constraint properties are read from the annotations applied by the
// cli/flag package.
func AllFlags(fs *pflag.FlagSet) []*Flag {
	var result []*Flag
	fs.VisitAll(func(f *pflag.Flag) {
		ft := ""
		if f.Value != nil {
			ft = f.Value.Type()
		}
		fv := Flag{
			Long:  f.Name,
			Short: f.Shorthand,
			Type:  ft,
			Group: annotation.Group(f),

			ExclusiveWith:   annotation.MutuallyExclusive(f),
			RequiredWith:    annotation.RequiredTogether(f),
			OneRequiredWith: annotation.OneRequired(f),
			Required:        annotation.IsRequired(f),
		}
		result = append(result, &fv)
	})
	slices.SortFunc(result, func(lhs, rhs *Flag) int {
		return strings.Compare(lhs.Long, rhs.Long)
	})
	return result
}

// LongFlags returns the long names of every flag registered in fs, sorted.
func LongFlags(fs *pflag.FlagSet) []string {
	all := AllFlags(fs)
	long := make([]string, 0, len(all))
	for _, f := range all {
		long = append(long, f.Long)
	}
	return long
}

// ShortFlags returns the shorthand names of every flag registered in fs, in
// long-name order.
func ShortFlags(fs *pflag.FlagSet) []string {
	all := AllFlags(fs)
	short := make([]string, 0, len(all))
	for _, f := range all {
		short = append(short, f.Short)
	}
	return short
}
