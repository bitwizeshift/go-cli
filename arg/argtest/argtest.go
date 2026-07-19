package argtest

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/internal/argdef"
)

// Parse parses args into cl, binding both the registered flags and the
// registered positional and unmatched arguments. It fails the test via
// [testing.TB.Fatalf] if parsing or binding returns an error.
func Parse(t testing.TB, cl *arg.CommandLine, args ...string) {
	t.Helper()

	ctx := context.Background()
	if err := cl.FlagSet().Parse(args); err != nil {
		t.Fatalf("Parse(...): unexpected error: %v", err)
	}
	if err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), cl.FlagSet().Args()); err != nil {
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
	reg := argdef.New()
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
	Usage string
}

// AllPositionals returns a [Positional] for every positional argument registered
// in cl, in registration order.
func AllPositionals(cl *arg.CommandLine) []*Positional {
	var result []*Positional
	for _, p := range argdef.Positionals((*argdef.CommandLine)(cl)) {
		result = append(result, &Positional{
			Index: p.Index,
			Name:  p.Name,
			Usage: p.Usage,
		})
	}
	return result
}

// Unmatched is a small wrapper around a registered unmatched-argument binding,
// for property-testing purposes.
type Unmatched struct {
	Name  string
	Usage string
}

// GetUnmatched returns the [Unmatched] binding registered in cl, or nil if cl
// has none.
func GetUnmatched(cl *arg.CommandLine) *Unmatched {
	u := argdef.GetUnmatched((*argdef.CommandLine)(cl))
	if u == nil {
		return nil
	}
	return &Unmatched{
		Name:  u.Name,
		Usage: u.Usage,
	}
}
