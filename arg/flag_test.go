package arg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

// addFlag constructs a flag named name bound to v and registers it on cl,
// returning the constructed [arg.FlagArg].
func addFlag[T any](cl *arg.CommandLine, name string, v *T, options ...arg.Option) *arg.FlagArg {
	f := arg.Flag(name, v, options...)
	cl.Add(f)
	return f
}

// addPositional constructs a positional argument bound to v at index and
// registers it on cl.
func addPositional[T any](cl *arg.CommandLine, name string, index int, v *T, options ...arg.Option) {
	cl.Add(arg.Positional(name, index, v, options...))
}

// addUnmatched constructs an unmatched-argument binding to out and registers it
// on cl.
func addUnmatched(cl *arg.CommandLine, out *[]string) {
	cl.Add(arg.Unmatched(out))
}

// newFlag constructs a string flag named name.
func newFlag(name string, options ...arg.Option) *arg.FlagArg {
	return arg.Flag(name, new(string), options...)
}

// newIntFlag constructs an int flag named name, to observe a differing reported
// type.
func newIntFlag(name string) *arg.FlagArg {
	return arg.Flag(name, new(int))
}

// newGroupedFlag constructs a string flag named name that belongs to the named
// display group.
func newGroupedFlag(name, group string) *arg.FlagArg {
	f := newFlag(name)
	arg.Group(group, f)
	return f
}

// newExclusiveFlag constructs a string flag named name that is mutually
// exclusive with a second flag.
func newExclusiveFlag(name string) *arg.FlagArg {
	f := newFlag(name)
	other := newFlag("other")
	arg.MarkMutuallyExclusive(f, other)
	return f
}

// newRequiredTogetherFlag constructs a string flag named name that is required
// together with a second flag.
func newRequiredTogetherFlag(name string) *arg.FlagArg {
	f := newFlag(name)
	other := newFlag("other")
	arg.MarkRequiredTogether(f, other)
	return f
}

// newOneRequiredFlag constructs a string flag named name of which at least one
// of it and a second flag is required.
func newOneRequiredFlag(name string) *arg.FlagArg {
	f := newFlag(name)
	other := newFlag("other")
	arg.MarkOneRequired(f, other)
	return f
}

// pflagName returns the long name of f, or an empty string if f is nil.
func pflagName(f *pflag.Flag) string {
	if f == nil {
		return ""
	}
	return f.Name
}

func TestFlagArg_Equal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		lhs  *arg.FlagArg
		rhs  *arg.FlagArg
		want bool
	}{
		{
			name: "SameConfiguration",
			lhs:  newFlag("name", arg.Shorthand("n"), arg.Usage("the name")),
			rhs:  newFlag("name", arg.Shorthand("n"), arg.Usage("the name")),
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
			lhs:  newFlag("name", arg.Shorthand("o")),
			rhs:  newFlag("name", arg.Shorthand("n")),
			want: false,
		},
		{
			name: "DifferentUsage",
			lhs:  newFlag("name", arg.Usage("something else")),
			rhs:  newFlag("name", arg.Usage("the name")),
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
			lhs:  newFlag("name", arg.Hidden()),
			rhs:  newFlag("name"),
			want: false,
		},
		{
			name: "DifferentRequired",
			lhs:  newFlag("name", arg.Required()),
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
			rhs:  &arg.FlagArg{},
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
				t.Errorf("FlagArg.Equal(...) = %t, want %t", got, want)
			}
		})
	}
}

func TestFlagArg_Flag(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		flag     *arg.FlagArg
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
			flag:     &arg.FlagArg{},
			wantNil:  true,
			wantName: "",
		},
		{
			name:     "RegisteredFlagExposesUnderlyingFlag",
			flag:     newFlag("name", arg.Shorthand("n")),
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
				t.Errorf("FlagArg.Flag() = nil %t, want %t", got, want)
			}
			if got, want := name, tc.wantName; !cmp.Equal(got, want) {
				t.Errorf("FlagArg.Flag().Name = %q, want %q", got, want)
			}
		})
	}
}

func TestMarkRequired(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	a := addFlag(cl, "a", new(bool))
	addFlag(cl, "b", new(bool))
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
	a := addFlag(cl, "a", new(bool))
	b := addFlag(cl, "b", new(bool))
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
	a := addFlag(cl, "a", new(bool))
	b := addFlag(cl, "b", new(bool))
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
	a := addFlag(cl, "a", new(bool))
	b := addFlag(cl, "b", new(bool))
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
