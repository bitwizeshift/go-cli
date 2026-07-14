package flag_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/flag/flagtest"
)

// newFlag registers a string flag named name in a fresh registry.
func newFlag(name string, options ...flag.Option) *flag.Flag {
	return flag.Add(flagtest.NewRegistry(), name, new(string), options...)
}

// newIntFlag registers an int flag named name in a fresh registry, to observe a
// differing reported type.
func newIntFlag(name string) *flag.Flag {
	return flag.Add(flagtest.NewRegistry(), name, new(int))
}

// newGroupedFlag registers a string flag named name that belongs to the named
// display group.
func newGroupedFlag(name, group string) *flag.Flag {
	f := newFlag(name)
	flag.AddToGroup(group, f)
	return f
}

// newExclusiveFlag registers a string flag named name that is mutually exclusive
// with a second flag in the same registry.
func newExclusiveFlag(name string) *flag.Flag {
	registry := flagtest.NewRegistry()
	f := flag.Add(registry, name, new(string))
	other := flag.Add(registry, "other", new(string))
	flag.MarkMutuallyExclusive(f, other)
	return f
}

// newRequiredTogetherFlag registers a string flag named name that is required
// together with a second flag in the same registry.
func newRequiredTogetherFlag(name string) *flag.Flag {
	registry := flagtest.NewRegistry()
	f := flag.Add(registry, name, new(string))
	other := flag.Add(registry, "other", new(string))
	flag.MarkRequiredTogether(f, other)
	return f
}

// newOneRequiredFlag registers a string flag named name of which at least one of
// it and a second flag in the same registry is required.
func newOneRequiredFlag(name string) *flag.Flag {
	registry := flagtest.NewRegistry()
	f := flag.Add(registry, name, new(string))
	other := flag.Add(registry, "other", new(string))
	flag.MarkOneRequired(f, other)
	return f
}

// pflagName returns the long name of f, or an empty string if f is nil.
func pflagName(f *pflag.Flag) string {
	if f == nil {
		return ""
	}
	return f.Name
}

func TestFlag_Equal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		lhs  *flag.Flag
		rhs  *flag.Flag
		want bool
	}{
		{
			name: "SameConfiguration",
			lhs:  newFlag("name", flag.Shorthand("n"), flag.Usage("the name")),
			rhs:  newFlag("name", flag.Shorthand("n"), flag.Usage("the name")),
			want: true,
		},
		{
			name: "DifferentName",
			lhs:  newFlag("other"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentShorthand",
			lhs:  newFlag("name", flag.Shorthand("o")),
			rhs:  newFlag("name", flag.Shorthand("n")),
			want: false,
		},
		{
			name: "DifferentUsage",
			lhs:  newFlag("name", flag.Usage("something else")),
			rhs:  newFlag("name", flag.Usage("the name")),
			want: false,
		},
		{
			name: "DifferentType",
			lhs:  newIntFlag("name"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentHidden",
			lhs:  newFlag("name", flag.Hidden()),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentRequired",
			lhs:  newFlag("name", flag.Required()),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentGroup",
			lhs:  newGroupedFlag("name", "Location Flags"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentMutuallyExclusiveWith",
			lhs:  newExclusiveFlag("name"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentRequiredWith",
			lhs:  newRequiredTogetherFlag("name"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentOneRequiredWith",
			lhs:  newOneRequiredFlag("name"),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "BothNil",
			lhs:  nil,
			rhs:  nil,
			want: true,
		},
		{
			name: "NilEqualsZeroFlag",
			lhs:  nil,
			rhs:  &flag.Flag{},
			want: true,
		},
		{
			name: "NilAndRegistered",
			lhs:  nil,
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "RegisteredAndNil",
			lhs:  newFlag("name"),
			rhs:  nil,
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.lhs

			// Act
			equal := sut.Equal(tc.rhs)

			// Assert
			if got, want := equal, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Flag.Equal(...) = %t, want %t", got, want)
			}
		})
	}
}

func TestFlag_Flag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		flag     *flag.Flag
		wantNil  bool
		wantName string
	}{
		{
			name:     "NilFlagHasNoUnderlyingFlag",
			flag:     nil,
			wantNil:  true,
			wantName: "",
		},
		{
			name:     "ZeroFlagHasNoUnderlyingFlag",
			flag:     &flag.Flag{},
			wantNil:  true,
			wantName: "",
		},
		{
			name:     "RegisteredFlagExposesUnderlyingFlag",
			flag:     newFlag("name", flag.Shorthand("n")),
			wantNil:  false,
			wantName: "name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := tc.flag

			// Act
			underlying := sut.Flag()
			isNil := underlying == nil
			name := pflagName(underlying)

			// Assert
			if got, want := isNil, tc.wantNil; !cmp.Equal(got, want) {
				t.Errorf("Flag.Flag() = nil %t, want %t", got, want)
			}
			if got, want := name, tc.wantName; !cmp.Equal(got, want) {
				t.Errorf("Flag.Flag().Name = %q, want %q", got, want)
			}
		})
	}
}

func TestRegistry_Flags(t *testing.T) {
	t.Parallel()

	// Arrange
	registry := flagtest.NewRegistry()
	flag.Add(registry, "verbose", new(bool), flag.Shorthand("v"))
	flag.Add(registry, "name", new(string))
	want := []*flag.Flag{
		newFlag("name"),
		flag.Add(flagtest.NewRegistry(), "verbose", new(bool), flag.Shorthand("v")),
	}

	// Act
	flags := registry.Flags()

	// Assert
	if got, want := flags, want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Registry.Flags() = %v, want %v\n%s", got, want, cmp.Diff(want, got))
	}
}
