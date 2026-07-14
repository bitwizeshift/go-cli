package flagtest

import (
	"slices"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/annotation"
	"github.com/bitwizeshift/go-cli/internal/flagreg"
	"github.com/spf13/pflag"
)

// Completion is the outcome of completing a flag: the values that were offered,
// and the behavior a shell is expected to fall back on.
//
// A nil [Completion] denotes a flag that offers no completion at all; all of its
// observers report zero-values.
type Completion struct {
	values    []string
	directive annotation.CompletionDirective
}

// Candidates returns the values offered for the completed word. It is empty
// unless the flag completes values.
func (c *Completion) Candidates() []string {
	if c == nil || c.directive != annotation.CompletionNoFileComp {
		return nil
	}
	return c.values
}

// FileExtensions returns the extensions, without a leading dot, that completed
// file names are filtered by. It is empty unless the flag completes files
// matching an extension.
func (c *Completion) FileExtensions() []string {
	if c == nil || c.directive != annotation.CompletionFilterFileExt {
		return nil
	}
	return c.values
}

// CompletesValues reports whether the flag completes with its own candidates,
// rather than with paths from the file-system.
func (c *Completion) CompletesValues() bool {
	return c != nil && c.directive == annotation.CompletionNoFileComp
}

// CompletesFiles reports whether the flag completes with file names, either any
// file or only the ones matching [Completion.FileExtensions].
func (c *Completion) CompletesFiles() bool {
	if c == nil {
		return false
	}
	return c.directive == annotation.CompletionDefault || c.directive == annotation.CompletionFilterFileExt
}

// CompletesDirs reports whether the flag completes with directory names.
func (c *Completion) CompletesDirs() bool {
	return c != nil && c.directive == annotation.CompletionFilterDirs
}

// CompleteFlag completes f with the partial word toComplete, returning the
// [Completion] it offers. It returns nil if f has no completion registered.
func CompleteFlag(f *pflag.Flag, toComplete string) *Completion {
	fn := annotation.GetCompletionFunc(f)
	if fn == nil {
		return nil
	}
	values, directive := fn(toComplete)
	return &Completion{
		values:    values,
		directive: directive,
	}
}

// Parse parses args into the flags registered in registry, failing the test via
// [testing.TB.Fatalf] if parsing returns an error.
func Parse(t testing.TB, registry *flag.Registry, args ...string) {
	t.Helper()

	fs := flagreg.Flags((*flagreg.Registry)(registry))
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

// NewRegistry returns a new [flag.Registry] that can be used for testing.
func NewRegistry() *flag.Registry {
	reg := flagreg.New()
	return (*flag.Registry)(reg)
}

// AllFlags returns a [Flag] for every flag registered in fs, sorted by long
// name. Constraint properties are read from the annotations applied by the
// cli/flag package.
func AllFlags(registry *flag.Registry) []*Flag {
	fs := flagreg.Flags((*flagreg.Registry)(registry))

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
