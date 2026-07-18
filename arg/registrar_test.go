package arg_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
)

// boolFlag is a [arg.Registrar] that registers a single bool flag identified
// by name.
type boolFlag struct {
	name string
}

func (b boolFlag) RegisterArgs(cl *arg.CommandLine) {
	var v bool
	addFlag(cl, b.name, &v)
}

var _ arg.Registrar = boolFlag{}

// container is a non-[arg.Registrar] struct that holds a [arg.Registrar]
// field, used to exercise recursive registration through pointers and
// interfaces.
type container struct {
	Flag boolFlag
}

// ifaceHolder holds an interface-typed field, used to exercise recursion
// through a [reflect.Interface] value whose concrete type is not itself a
// [arg.Registrar].
type ifaceHolder struct {
	Value any
}

// unexportedField holds an unexported [arg.Registrar] field, which must be
// skipped because its value cannot be read reflectively.
type unexportedField struct {
	hidden boolFlag
}

// tagged holds a [arg.Registrar] field annotated to be ignored via a struct
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

// pair holds two [arg.Registrar] pointer fields, used to place the same
// instance at more than one position within a tree.
type pair struct {
	First  *boolFlag
	Second *boolFlag
}

// ifacePair holds two interface-typed fields, used to reach the same pointer
// instance through interface values rather than concrete pointer fields.
type ifacePair struct {
	First  any
	Second any
}

// manualParent is a [arg.Registrar] whose RegisterArgs re-enters registration
// for the same child twice, exercising deduplication across the RegisterArgs
// boundary.
type manualParent struct {
	Child *boolFlag
}

func (m *manualParent) RegisterArgs(cl *arg.CommandLine) {
	arg.Register(cl, m.Child)
	arg.Register(cl, m.Child)
}

var _ arg.Registrar = (*manualParent)(nil)

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
			cl := argtest.NewCommandLine()

			// Act
			arg.Register(cl, tc.v)

			// Assert
			names := argtest.LongFlags(cl)
			opts := cmp.Options{cmpopts.SortSlices(strings.Compare), cmpopts.EquateEmpty()}
			if got, want := names, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("Register(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got, opts...))
			}
		})
	}
}

func TestRegister_UniqueInstances(t *testing.T) {
	t.Parallel()

	shared := &boolFlag{name: "shared"}
	manualChild := &boolFlag{name: "child"}
	distinctA := &boolFlag{name: "a"}
	distinctB := &boolFlag{name: "b"}

	testCases := []struct {
		name string
		v    any
		want []string
	}{
		{
			name: "SamePointerInTwoStructFields",
			v:    &pair{First: shared, Second: shared},
			want: []string{"shared"},
		},
		{
			name: "SamePointerTwiceInSlice",
			v:    []*boolFlag{shared, shared},
			want: []string{"shared"},
		},
		{
			name: "SamePointerThroughInterfaceFields",
			v:    &ifacePair{First: shared, Second: shared},
			want: []string{"shared"},
		},
		{
			name: "ManualReRegistrationDeduplicated",
			v:    &manualParent{Child: manualChild},
			want: []string{"child"},
		},
		{
			name: "DistinctPointersBothRegister",
			v:    &pair{First: distinctA, Second: distinctB},
			want: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()

			// Act
			arg.Register(cl, tc.v)

			// Assert
			names := argtest.LongFlags(cl)
			opts := cmp.Options{cmpopts.SortSlices(strings.Compare), cmpopts.EquateEmpty()}
			if got, want := names, tc.want; !cmp.Equal(got, want, opts...) {
				t.Errorf("Register(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got, opts...))
			}
		})
	}
}
