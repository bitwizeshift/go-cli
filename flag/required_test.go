package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
)

// newFlagSet returns a flag set with two boolean flags "a" and "b".
func newFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("a", false, "")
	fs.Bool("b", false, "")
	return fs
}

func TestMarkRequired(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := newFlagSet()
	flag.MarkRequired(fs.Lookup("a"))

	// Act
	flags := flagtest.AllFlags(fs)

	// Assert
	want := []*flagtest.Flag{
		{Long: "a", Type: "bool", Required: true},
		{Long: "b", Type: "bool"},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestMarkRequiredTogether(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := newFlagSet()
	flag.MarkRequiredTogether(fs.Lookup("a"), fs.Lookup("b"))

	// Act
	flags := flagtest.AllFlags(fs)

	// Assert
	want := []*flagtest.Flag{
		{Long: "a", Type: "bool", RequiredWith: []string{"a", "b"}},
		{Long: "b", Type: "bool", RequiredWith: []string{"a", "b"}},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestMarkMutuallyExclusive(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := newFlagSet()
	flag.MarkMutuallyExclusive(fs.Lookup("a"), fs.Lookup("b"))

	// Act
	flags := flagtest.AllFlags(fs)

	// Assert
	want := []*flagtest.Flag{
		{Long: "a", Type: "bool", ExclusiveWith: []string{"a", "b"}},
		{Long: "b", Type: "bool", ExclusiveWith: []string{"a", "b"}},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}

func TestMarkOneRequired(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := newFlagSet()
	flag.MarkOneRequired(fs.Lookup("a"), fs.Lookup("b"))

	// Act
	flags := flagtest.AllFlags(fs)

	// Assert
	want := []*flagtest.Flag{
		{Long: "a", Type: "bool", OneRequiredWith: []string{"a", "b"}},
		{Long: "b", Type: "bool", OneRequiredWith: []string{"a", "b"}},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
