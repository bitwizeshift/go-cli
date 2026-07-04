package flag_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
)

// boolFlag is a [flag.Registrar] that registers a single bool flag identified
// by name.
type boolFlag struct {
	name string
}

func (b boolFlag) RegisterFlags(fs *pflag.FlagSet) {
	fs.Bool(b.name, false, "")
}

var _ flag.Registrar = boolFlag{}

// container is a non-[flag.Registrar] struct that holds a [flag.Registrar]
// field, used to exercise recursive registration through pointers and
// interfaces.
type container struct {
	Flag boolFlag
}

// ifaceHolder holds an interface-typed field, used to exercise recursion
// through a [reflect.Interface] value whose concrete type is not itself a
// [flag.Registrar].
type ifaceHolder struct {
	Value any
}

// unexportedField holds an unexported [flag.Registrar] field, which must be
// skipped because its value cannot be read reflectively.
type unexportedField struct {
	hidden boolFlag
}

// tagged holds a [flag.Registrar] field annotated to be ignored via a struct
// tag.
type tagged struct {
	Flag boolFlag `flag:"ignore"`
}

// mixed holds a registered field, a dash-ignored field, and a non-registrar
// scalar field to exercise the visible, ignored, and inert branches together.
type mixed struct {
	Enabled boolFlag
	Skipped boolFlag `flag:"-"`
	Count   int
}

func TestRegister(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want []string
	}{
		{
			name: "DirectRegistrar",
			v:    boolFlag{name: "direct"},
			want: []string{"direct"},
		},
		{
			name: "PointerToStructRecurses",
			v:    &container{Flag: boolFlag{name: "ptr"}},
			want: []string{"ptr"},
		},
		{
			name: "InterfaceFieldRecurses",
			v:    ifaceHolder{Value: container{Flag: boolFlag{name: "iface"}}},
			want: []string{"iface"},
		},
		{
			name: "UnexportedFieldSkipped",
			v:    unexportedField{hidden: boolFlag{name: "hidden"}},
			want: nil,
		},
		{
			name: "IgnoreTagSkipped",
			v:    tagged{Flag: boolFlag{name: "ignored"}},
			want: nil,
		},
		{
			name: "MixedStructRegistersOnlyVisibleNonIgnoredFields",
			v: mixed{
				Enabled: boolFlag{name: "enabled"},
				Skipped: boolFlag{name: "skip"},
				Count:   5,
			},
			want: []string{"enabled"},
		},
		{
			name: "SliceRegistersEachElement",
			v:    []boolFlag{{name: "s0"}, {name: "s1"}},
			want: []string{"s0", "s1"},
		},
		{
			name: "ArrayRegistersEachElement",
			v:    [2]boolFlag{{name: "a0"}, {name: "a1"}},
			want: []string{"a0", "a1"},
		},
		{
			name: "MapRegistersValues",
			v:    map[string]boolFlag{"key": {name: "mapval"}},
			want: []string{"mapval"},
		},
		{
			name: "MapRegistersKeys",
			v:    map[boolFlag]int{{name: "mapkey"}: 0},
			want: []string{"mapkey"},
		},
		{
			name: "NilMapRegistersNothing",
			v:    map[string]boolFlag(nil),
			want: nil,
		},
		{
			name: "NonRegistrarScalarRegistersNothing",
			v:    42,
			want: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

			// Act
			flag.Register(fs, tc.v)

			// Assert
			names := flagNames(fs)
			opts := cmp.Options{cmpopts.SortSlices(strings.Compare), cmpopts.EquateEmpty()}
			if got, want := names, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("Register(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got, opts...))
			}
		})
	}
}

// flagNames returns the names of all flags registered in fs, sorted
// lexicographically by [pflag.FlagSet.VisitAll].
func flagNames(fs *pflag.FlagSet) []string {
	var names []string
	fs.VisitAll(func(f *pflag.Flag) {
		names = append(names, f.Name)
	})
	return names
}
