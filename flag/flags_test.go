package flag_test

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"

	"github.com/bitwizeshift/go-cli/flag"
	"github.com/bitwizeshift/go-cli/internal/annotation"
)

// opCode is a defined multi-word type used to exercise the default kebab-case
// type-name derivation, including the package prefix.
type opCode uint8

// toggle is a defined bool type used to confirm that defined types do not
// receive the bool bare-flag default.
type toggle bool

// stringList is a defined slice type used to confirm that defined slice types
// do not accumulate across repeated occurrences.
type stringList []string

// errDecode is returned by failing test decoders to observe error propagation.
var errDecode = errors.New("decode failed")

// yell is a string decoder that upper-cases its input.
func yell(data []byte) (string, error) { return strings.ToUpper(string(data)), nil }

// failString is a string decoder that always fails with errDecode.
func failString([]byte) (string, error) { return "", errDecode }

// parseHexInt is an int decoder that reads a base-16 integer without a prefix.
func parseHexInt(data []byte) (int, error) {
	n, err := strconv.ParseInt(string(data), 16, 0)
	return int(n), err
}

// setEach applies each value to v in order, mirroring one flag occurrence per
// value, and returns the first error encountered.
func setEach(v pflag.Value, values []string) error {
	for _, s := range values {
		if err := v.Set(s); err != nil {
			return err
		}
	}
	return nil
}

// flagInfo captures the observable properties of a registered flag for
// full-structure comparison.
type flagInfo struct {
	Short string
	Type  string
	Usage string
	NoOpt string
}

// infoOf reads the observable properties of f into a [flagInfo].
func infoOf(f *pflag.Flag) flagInfo {
	return flagInfo{
		Short: f.Shorthand,
		Type:  f.Value.Type(),
		Usage: f.Usage,
		NoOpt: f.NoOptDefVal,
	}
}

func TestAdd_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
		sets    []string
		want    string
		wantErr error
	}{
		{
			name: "SetsValue",
			sets: []string{"hello"},
			want: "hello",
		},
		{
			name:    "RejectsSecondValue",
			sets:    []string{"a", "b"},
			want:    "a",
			wantErr: flag.ErrAlreadySet,
		},
		{
			name:    "UnmarshalWithOverridesDecoder",
			options: []flag.Option{flag.UnmarshalWith(yell)},
			sets:    []string{"hi"},
			want:    "HI",
		},
		{
			name:    "UnmarshalWithPropagatesError",
			options: []flag.Option{flag.UnmarshalWith(failString)},
			sets:    []string{"x"},
			want:    "",
			wantErr: errDecode,
		},
		{
			name:    "UnmarshalWithTypeMismatch",
			options: []flag.Option{flag.UnmarshalWith(parseHexInt)},
			sets:    []string{"ff"},
			want:    "",
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst string

			// Act
			f := flag.Add(fs, "value", &dst, tc.options...)
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) value = %v, want %v", got, want)
			}
		})
	}
}

func TestAdd_Int(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
		sets    []string
		want    int
		wantErr error
	}{
		{
			name: "Decimal",
			sets: []string{"42"},
			want: 42,
		},
		{
			name:    "DecodeError",
			sets:    []string{"abc"},
			want:    0,
			wantErr: strconv.ErrSyntax,
		},
		{
			name:    "HexViaUnmarshalWith",
			options: []flag.Option{flag.UnmarshalWith(parseHexInt)},
			sets:    []string{"ff"},
			want:    255,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst int

			// Act
			f := flag.Add(fs, "n", &dst, tc.options...)
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) value = %v, want %v", got, want)
			}
		})
	}
}

func TestAdd_Bool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		sets    []string
		want    bool
		wantErr error
	}{
		{
			name: "True",
			sets: []string{"true"},
			want: true,
		},
		{
			name: "False",
			sets: []string{"false"},
			want: false,
		},
		{
			name:    "RejectsSecondValue",
			sets:    []string{"true", "false"},
			want:    true,
			wantErr: flag.ErrAlreadySet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst bool

			// Act
			f := flag.Add(fs, "verbose", &dst)
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) value = %v, want %v", got, want)
			}
		})
	}
}

func TestAdd_Slice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		sets []string
		want []int
	}{
		{
			name: "CommaSeparated",
			sets: []string{"1,2"},
			want: []int{1, 2},
		},
		{
			name: "RepeatedAccumulates",
			sets: []string{"1", "2"},
			want: []int{1, 2},
		},
		{
			name: "CommaSeparatedThenRepeated",
			sets: []string{"1,2", "3"},
			want: []int{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst []int

			// Act
			f := flag.Add(fs, "n", &dst)
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) value = %v, want %v", got, want)
			}
		})
	}
}

func TestAdd_DefinedSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		sets    []string
		want    stringList
		wantErr error
	}{
		{
			name: "DecodesCommaSeparated",
			sets: []string{"a,b"},
			want: stringList{"a", "b"},
		},
		{
			name:    "RejectsSecondValue",
			sets:    []string{"a", "b"},
			want:    stringList{"a"},
			wantErr: flag.ErrAlreadySet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst stringList

			// Act
			f := flag.Add(fs, "n", &dst)
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) value = %v, want %v", got, want)
			}
		})
	}
}

func TestAdd_Options(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
		want    flagInfo
	}{
		{
			name: "Defaults",
			want: flagInfo{Type: "string"},
		},
		{
			name:    "Shorthand",
			options: []flag.Option{flag.Shorthand("v")},
			want:    flagInfo{Short: "v", Type: "string"},
		},
		{
			name:    "TypeOverride",
			options: []flag.Option{flag.Type("custom")},
			want:    flagInfo{Type: "custom"},
		},
		{
			name:    "Usage",
			options: []flag.Option{flag.Usage("the value")},
			want:    flagInfo{Type: "string", Usage: "the value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst string

			// Act
			f := flag.Add(fs, "value", &dst, tc.options...)
			info := infoOf(f)

			// Assert
			if got, want := info, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) flag = %+v, want %+v", got, want)
			}
		})
	}
}

func TestAdd_TypeNameAndBareFlag(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
	var opDst opCode
	var boolDst bool
	var toggleDst toggle

	// Act
	opFlag := flag.Add(fs, "op", &opDst)
	boolFlag := flag.Add(fs, "bool", &boolDst)
	toggleFlag := flag.Add(fs, "toggle", &toggleDst)
	infos := []flagInfo{infoOf(opFlag), infoOf(boolFlag), infoOf(toggleFlag)}

	// Assert
	want := []flagInfo{
		{Type: "flag-test-op-code"},
		{Type: "bool", NoOpt: "true"},
		{Type: "flag-test-toggle"},
	}
	if got, want := infos, want; !cmp.Equal(got, want) {
		t.Errorf("Add(...) flags = %+v, want %+v", got, want)
	}
}

func TestAdd_NilPointerRendersEmptyString(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
	var dst *int
	f := flag.Add(fs, "n", &dst)

	// Act
	str := f.Value.String()

	// Assert
	if got, want := str, ""; !cmp.Equal(got, want) {
		t.Errorf("Add(...) string = %q, want %q", got, want)
	}
}

func TestAddCallback(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		sets    []string
		want    []int
		wantErr error
	}{
		{
			name: "InvokedPerValue",
			sets: []string{"5"},
			want: []int{5},
		},
		{
			name: "InvokedOncePerOccurrence",
			sets: []string{"1", "2"},
			want: []int{1, 2},
		},
		{
			name: "NotInvokedWhenAbsent",
			sets: nil,
			want: nil,
		},
		{
			name:    "DecodeErrorSkipsCallback",
			sets:    []string{"abc"},
			want:    nil,
			wantErr: strconv.ErrSyntax,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var seen []int
			cb := func(v int) error {
				seen = append(seen, v)
				return nil
			}
			f := flag.AddCallback(fs, "n", cb)

			// Act
			err := setEach(f.Value, tc.sets)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Set(...) error = %v, want %v", got, want)
			}
			if got, want := seen, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("AddCallback(...) invocations = %v, want %v", got, want)
			}
		})
	}
}

func TestAddCallback_BoolBareInvokesTrue(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
	var seen []bool
	cb := func(v bool) error {
		seen = append(seen, v)
		return nil
	}
	f := flag.AddCallback(fs, "verbose", cb)

	// Act
	err := f.Value.Set("true")
	info := infoOf(f)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Set(...) error = %v, want %v", got, want)
	}
	if got, want := info, (flagInfo{Type: "bool", NoOpt: "true"}); !cmp.Equal(got, want) {
		t.Errorf("AddCallback(...) flag = %+v, want %+v", got, want)
	}
	if got, want := seen, []bool{true}; !cmp.Equal(got, want) {
		t.Errorf("AddCallback(...) invocations = %v, want %v", got, want)
	}
}

func TestAddCallback_ErrorPropagates(t *testing.T) {
	t.Parallel()

	// Arrange
	fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
	cbErr := errors.New("callback failed")
	cb := func(int) error { return cbErr }
	f := flag.AddCallback(fs, "n", cb)

	// Act
	err := f.Value.Set("5")

	// Assert
	if got, want := err, cbErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Set(...) error = %v, want %v", got, want)
	}
}

func TestHidden(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
		want    bool
	}{
		{
			name:    "HiddenOptionMarksFlagHidden",
			options: []flag.Option{flag.Hidden()},
			want:    true,
		},
		{
			name:    "DefaultLeavesFlagVisible",
			options: nil,
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst string

			// Act
			f := flag.Add(fs, "flag", &dst, tc.options...)

			// Assert
			if got, want := f.Hidden, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) hidden = %t, want %t", got, want)
			}
		})
	}
}

func TestRequired(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []flag.Option
		want    bool
	}{
		{
			name:    "RequiredOptionMarksFlagRequired",
			options: []flag.Option{flag.Required()},
			want:    true,
		},
		{
			name:    "DefaultLeavesFlagOptional",
			options: nil,
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fs := flag.NewRegistry(pflag.NewFlagSet("test", pflag.ContinueOnError))
			var dst string

			// Act
			f := flag.Add(fs, "flag", &dst, tc.options...)

			// Assert
			if got, want := annotation.IsRequired(f), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Add(...) required = %t, want %t", got, want)
			}
		})
	}
}

func TestDefaultFromEnv(t *testing.T) {
	// Arrange
	pfs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs := flag.NewRegistry(pfs)
	var dst string
	flag.Add(fs, "flag", &dst, flag.DefaultFromEnv("FLAG_ENV"))
	t.Setenv("FLAG_ENV", "from-env")
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, pfs)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
	}
	if got, want := dst, "from-env"; !cmp.Equal(got, want) {
		t.Errorf("flag value = %q, want %q", got, want)
	}
}

func TestDefaultFromFunc(t *testing.T) {
	t.Parallel()

	// Arrange
	pfs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs := flag.NewRegistry(pfs)
	var dst string
	fn := func(context.Context) (string, error) { return "from-func", nil }
	flag.Add(fs, "flag", &dst, flag.DefaultFromFunc(fn))
	ctx := context.Background()

	// Act
	err := annotation.SetFlagFallbacks(ctx, pfs)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("SetFlagFallbacks(...) = %v, want %v", got, want)
	}
	if got, want := dst, "from-func"; !cmp.Equal(got, want) {
		t.Errorf("flag value = %q, want %q", got, want)
	}
}
