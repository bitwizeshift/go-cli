package arg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

func TestMarkRequired(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	a := arg.AddFlag(cl, "a", new(bool))
	arg.AddFlag(cl, "b", new(bool))
	arg.MarkRequired(a)

	// Act
	flags := argtest.AllFlags(cl)

	// Assert
	want := []*argtest.Flag{
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
	cl := argtest.NewCommandLine()
	a := arg.AddFlag(cl, "a", new(bool))
	b := arg.AddFlag(cl, "b", new(bool))
	arg.MarkRequiredTogether(a, b)

	// Act
	flags := argtest.AllFlags(cl)

	// Assert
	want := []*argtest.Flag{
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
	cl := argtest.NewCommandLine()
	a := arg.AddFlag(cl, "a", new(bool))
	b := arg.AddFlag(cl, "b", new(bool))
	arg.MarkMutuallyExclusive(a, b)

	// Act
	flags := argtest.AllFlags(cl)

	// Assert
	want := []*argtest.Flag{
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
	cl := argtest.NewCommandLine()
	a := arg.AddFlag(cl, "a", new(bool))
	b := arg.AddFlag(cl, "b", new(bool))
	arg.MarkOneRequired(a, b)

	// Act
	flags := argtest.AllFlags(cl)

	// Assert
	want := []*argtest.Flag{
		{Long: "a", Type: "bool", OneRequiredWith: []string{"a", "b"}},
		{Long: "b", Type: "bool", OneRequiredWith: []string{"a", "b"}},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
