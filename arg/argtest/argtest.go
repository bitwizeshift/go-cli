package argtest

import (
	"slices"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argreg"
)

// Parse parses args into cl, binding both the registered flags and the
// registered positional and unmatched arguments. It fails the test via
// [testing.TB.Fatalf] if parsing or binding returns an error.
func Parse(t testing.TB, cl *arg.CommandLine, args ...string) {
	t.Helper()

	if err := cl.FlagSet().Parse(args); err != nil {
		t.Fatalf("Parse(...): unexpected error: %v", err)
	}
	if err := argreg.Bind((*argreg.CommandLine)(cl), cl.FlagSet().Args()); err != nil {
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

// NewCommandLine returns a new [arg.CommandLine] that can be used for testing.
func NewCommandLine() *arg.CommandLine {
	reg := argreg.New()
	return (*arg.CommandLine)(reg)
}

// AllFlags returns a [Flag] for every flag registered in cl, sorted by
// long name.
func AllFlags(cl *arg.CommandLine) []*Flag {
	var result []*Flag
	for _, f := range cl.Flags() {
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
func LongFlags(cl *arg.CommandLine) []string {
	all := AllFlags(cl)
	long := make([]string, 0, len(all))
	for _, f := range all {
		long = append(long, f.Long)
	}
	return long
}

// ShortFlags returns the shorthand names of every flag registered in fs, in
// long-name order.
func ShortFlags(cl *arg.CommandLine) []string {
	all := AllFlags(cl)
	short := make([]string, 0, len(all))
	for _, f := range all {
		short = append(short, f.Short)
	}
	return short
}

// Positional is a small wrapper around a registered positional argument, for
// property-testing purposes.
type Positional struct {
	Index int
	Name  string
	Type  string
	Usage string
}

// AllPositionals returns a [Positional] for every positional argument registered
// in cl, in registration order.
func AllPositionals(cl *arg.CommandLine) []*Positional {
	var result []*Positional
	for _, p := range argreg.Positionals((*argreg.CommandLine)(cl)) {
		result = append(result, &Positional{
			Index: p.Index,
			Name:  p.Name,
			Type:  p.Type,
			Usage: p.Usage,
		})
	}
	return result
}
