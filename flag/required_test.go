package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
)

func TestMarkRequired(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	a := flag.Add(registry, "a", new(bool))
	flag.Add(registry, "b", new(bool))
	flag.MarkRequired(a)

	// Act
	flags := flagtest.AllFlags(registry)

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
	registry := flagtest.NewRegistry()
	a := flag.Add(registry, "a", new(bool))
	b := flag.Add(registry, "b", new(bool))
	flag.MarkRequiredTogether(a, b)

	// Act
	flags := flagtest.AllFlags(registry)

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
	registry := flagtest.NewRegistry()
	a := flag.Add(registry, "a", new(bool))
	b := flag.Add(registry, "b", new(bool))
	flag.MarkMutuallyExclusive(a, b)

	// Act
	flags := flagtest.AllFlags(registry)

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
	registry := flagtest.NewRegistry()
	a := flag.Add(registry, "a", new(bool))
	b := flag.Add(registry, "b", new(bool))
	flag.MarkOneRequired(a, b)

	// Act
	flags := flagtest.AllFlags(registry)

	// Assert
	want := []*flagtest.Flag{
		{Long: "a", Type: "bool", OneRequiredWith: []string{"a", "b"}},
		{Long: "b", Type: "bool", OneRequiredWith: []string{"a", "b"}},
	}
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("AllFlags(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
