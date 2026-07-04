package flag_test

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/flag"
)

// readTarget returns the value that a successful [flag.Unmarshal] wrote into
// its target by fully dereferencing v. Targets that cannot receive a value --
// non-pointers and nil pointers -- normalize to nil.
func readTarget(t testing.TB, v any) any {
	t.Helper()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return nil
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}

// recordingUnmarshaler records the bytes passed to UnmarshalFlag.
type recordingUnmarshaler struct {
	Data string
}

func (r *recordingUnmarshaler) UnmarshalFlag(data []byte) error {
	r.Data = string(data)
	return nil
}

// errUnmarshaler always reports errUnmarshal from UnmarshalFlag.
type errUnmarshaler struct {
	Err error
}

func (u errUnmarshaler) UnmarshalFlag([]byte) error {
	return u.Err
}

// recordingTextUnmarshaler records the bytes passed to UnmarshalText.
type recordingTextUnmarshaler struct {
	Data string
}

func (r *recordingTextUnmarshaler) UnmarshalText(data []byte) error {
	r.Data = string(data)
	return nil
}

// dualUnmarshaler implements both [flag.Unmarshaler] and
// [encoding.TextUnmarshaler], recording which method was invoked.
type dualUnmarshaler struct {
	Used string
}

func (d *dualUnmarshaler) UnmarshalFlag([]byte) error {
	d.Used = "flag"
	return nil
}

func (d *dualUnmarshaler) UnmarshalText([]byte) error {
	d.Used = "text"
	return nil
}

var errFailUnmarshal = errors.New("unmarshal failed")

// failUnmarshaler fails UnmarshalFlag unconditionally with errFailUnmarshal.
type failUnmarshaler struct{}

func (*failUnmarshaler) UnmarshalFlag([]byte) error {
	return errFailUnmarshal
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")
	testCases := []struct {
		name    string
		target  any
		data    string
		want    any
		wantErr error
	}{
		{
			name:   "UnmarshalerIsPreferred",
			target: &recordingUnmarshaler{},
			data:   "payload",
			want:   recordingUnmarshaler{Data: "payload"},
		},
		{
			name:    "UnmarshalerErrorPropagates",
			target:  errUnmarshaler{Err: testErr},
			data:    "payload",
			wantErr: testErr,
		},
		{
			name:   "TextUnmarshalerIsUsed",
			target: &recordingTextUnmarshaler{},
			data:   "payload",
			want:   recordingTextUnmarshaler{Data: "payload"},
		},
		{
			name:   "UnmarshalerTakesPrecedenceOverText",
			target: &dualUnmarshaler{},
			data:   "payload",
			want:   dualUnmarshaler{Used: "flag"},
		},
		{
			name:   "StringVerbatim",
			target: new(string),
			data:   "hello world",
			want:   "hello world",
		},
		{
			name:   "StringEmpty",
			target: new(string),
			data:   "",
			want:   "",
		},
		{
			name:   "BoolTrue",
			target: new(bool),
			data:   "true",
			want:   true,
		},
		{
			name:   "BoolFalse",
			target: new(bool),
			data:   "false",
			want:   false,
		},
		{
			name:    "BoolRejectsNumeric",
			target:  new(bool),
			data:    "42",
			want:    false,
			wantErr: strconv.ErrSyntax,
		},
		{
			name:   "IntDecimal",
			target: new(int),
			data:   "42",
			want:   42,
		},
		{
			name:   "IntNegative",
			target: new(int),
			data:   "-42",
			want:   -42,
		},
		{
			name:   "IntHex",
			target: new(int),
			data:   "0xff",
			want:   255,
		},
		{
			name:   "IntBinary",
			target: new(int),
			data:   "0b101",
			want:   5,
		},
		{
			name:   "IntOctal",
			target: new(int),
			data:   "0o17",
			want:   15,
		},
		{
			name:   "Int32Sized",
			target: new(int32),
			data:   "1000",
			want:   int32(1000),
		},
		{
			name:   "Int64Sized",
			target: new(int64),
			data:   "9999999999",
			want:   int64(9999999999),
		},
		{
			name:    "Int8Overflow",
			target:  new(int8),
			data:    "128",
			want:    int8(0),
			wantErr: strconv.ErrRange,
		},
		{
			name:    "IntSyntax",
			target:  new(int),
			data:    "abc",
			want:    0,
			wantErr: strconv.ErrSyntax,
		},
		{
			name:   "UintHex",
			target: new(uint),
			data:   "0xFF",
			want:   uint(255),
		},
		{
			name:   "Uint8Sized",
			target: new(uint8),
			data:   "255",
			want:   uint8(255),
		},
		{
			name:    "UintRejectsNegative",
			target:  new(uint),
			data:    "-1",
			want:    uint(0),
			wantErr: strconv.ErrSyntax,
		},
		{
			name:   "Float64",
			target: new(float64),
			data:   "3.14",
			want:   3.14,
		},
		{
			name:   "Float32",
			target: new(float32),
			data:   "1.5",
			want:   float32(1.5),
		},
		{
			name:    "FloatSyntax",
			target:  new(float64),
			data:    "not-a-number",
			want:    float64(0),
			wantErr: strconv.ErrSyntax,
		},
		{
			name:   "Duration",
			target: new(time.Duration),
			data:   "1h30m",
			want:   90 * time.Minute,
		},
		{
			name:    "DurationInvalid",
			target:  new(time.Duration),
			data:    "1parsec",
			want:    time.Duration(0),
			wantErr: cmpopts.AnyError,
		},
		{
			name:   "NilPointerAllocatedOnce",
			target: new(*int),
			data:   "7",
			want:   7,
		},
		{
			name:   "NilPointerAllocatedRecursively",
			target: new(**int),
			data:   "9",
			want:   9,
		},
		{
			name:    "NilPointerNotAllocatedOnError",
			target:  new(*int),
			data:    "abc",
			want:    nil,
			wantErr: strconv.ErrSyntax,
		},
		{
			name:   "NilUnmarshalerPointerAllocatedOnSuccess",
			target: new(*recordingUnmarshaler),
			data:   "payload",
			want:   recordingUnmarshaler{Data: "payload"},
		},
		{
			name:    "NilUnmarshalerPointerNotAllocatedOnError",
			target:  new(*failUnmarshaler),
			data:    "payload",
			want:    nil,
			wantErr: errFailUnmarshal,
		},
		{
			name:    "UnsupportedType",
			target:  new(map[string]int),
			data:    "x",
			want:    map[string]int(nil),
			wantErr: flag.ErrUnsupportedType,
		},
		{
			name:   "SliceString",
			target: new([]string),
			data:   "a,b,c",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "SliceStringEmpty",
			target: new([]string),
			data:   "",
			want:   []string{},
		},
		{
			name:   "SliceStringSingle",
			target: new([]string),
			data:   "a",
			want:   []string{"a"},
		},
		{
			name:   "SliceStringQuoted",
			target: new([]string),
			data:   `"a,b",c`,
			want:   []string{"a,b", "c"},
		},
		{
			name:   "SliceInt",
			target: new([]int),
			data:   "1,2,3",
			want:   []int{1, 2, 3},
		},
		{
			name:    "SliceIntElementError",
			target:  new([]int),
			data:    "1,abc",
			want:    []int(nil),
			wantErr: strconv.ErrSyntax,
		},
		{
			name:    "SliceMalformedCSV",
			target:  new([]string),
			data:    `a"b`,
			want:    []string(nil),
			wantErr: cmpopts.AnyError,
		},
		{
			name:   "SliceUnmarshaler",
			target: new([]recordingUnmarshaler),
			data:   "x,y",
			want:   []recordingUnmarshaler{{Data: "x"}, {Data: "y"}},
		},
		{
			name:    "SliceUnsupportedElement",
			target:  new([]complex128),
			data:    "1",
			want:    []complex128(nil),
			wantErr: flag.ErrUnsupportedType,
		},
		{
			name:    "NonPointerTarget",
			target:  42,
			data:    "1",
			wantErr: flag.ErrInvalidTarget,
		},
		{
			name:    "NilPointerTarget",
			target:  (*int)(nil),
			data:    "1",
			wantErr: flag.ErrInvalidTarget,
		},
		{
			name:    "NilUnmarshalerTarget",
			target:  (*recordingUnmarshaler)(nil),
			data:    "payload",
			wantErr: flag.ErrInvalidTarget,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			data := []byte(tc.data)

			// Act
			err := flag.Unmarshal(tc.target, data)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Unmarshal(...) error = %v, want %v", got, want)
			}
			value := readTarget(t, tc.target)
			if got, want := value, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Unmarshal(...) = %v, want %v", got, want)
			}
		})
	}
}
