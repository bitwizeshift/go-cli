package ask_test

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/prompt/internal/ask"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// upperText is a TextUnmarshaler that upper-cases the text it is given.
type upperText struct {
	Value string
}

func (u *upperText) UnmarshalText(text []byte) error {
	u.Value = strings.ToUpper(string(text))
	return nil
}

// failText is a TextUnmarshaler that always fails with errText.
type failText struct{}

var errText = errors.New("bad text")

func (failText) UnmarshalText([]byte) error {
	return errText
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		text    string
		target  any
		want    any
		wantErr error
	}{
		{
			name:   "String",
			text:   "hello",
			target: new(string),
			want:   "hello",
		}, {
			name:   "BoolTrue",
			text:   "true",
			target: new(bool),
			want:   true,
		}, {
			name:    "BoolInvalid",
			text:    "maybe",
			target:  new(bool),
			want:    false,
			wantErr: strconv.ErrSyntax,
		}, {
			name:   "Int",
			text:   "-42",
			target: new(int),
			want:   -42,
		}, {
			name:   "Int8",
			text:   "127",
			target: new(int8),
			want:   int8(127),
		}, {
			name:    "Int8Overflow",
			text:    "128",
			target:  new(int8),
			want:    int8(0),
			wantErr: strconv.ErrRange,
		}, {
			name:    "IntInvalid",
			text:    "notanumber",
			target:  new(int),
			want:    0,
			wantErr: strconv.ErrSyntax,
		}, {
			name:   "Uint",
			text:   "42",
			target: new(uint),
			want:   uint(42),
		}, {
			name:    "UintNegative",
			text:    "-1",
			target:  new(uint16),
			want:    uint16(0),
			wantErr: strconv.ErrSyntax,
		}, {
			name:   "Float64",
			text:   "3.14",
			target: new(float64),
			want:   3.14,
		}, {
			name:   "Float32",
			text:   "1.5",
			target: new(float32),
			want:   float32(1.5),
		}, {
			name:    "FloatInvalid",
			text:    "pi",
			target:  new(float64),
			want:    float64(0),
			wantErr: strconv.ErrSyntax,
		}, {
			name:   "Duration",
			text:   "1500ms",
			target: new(time.Duration),
			want:   1500 * time.Millisecond,
		}, {
			name:    "DurationInvalid",
			text:    "soon",
			target:  new(time.Duration),
			want:    time.Duration(0),
			wantErr: cmpopts.AnyError,
		}, {
			name:   "TextUnmarshaler",
			text:   "hello",
			target: &upperText{},
			want:   upperText{Value: "HELLO"},
		}, {
			name:    "TextUnmarshalerError",
			text:    "anything",
			target:  &failText{},
			want:    failText{},
			wantErr: errText,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			err := ask.Unmarshal(tc.text, tc.target)
			value := reflect.ValueOf(tc.target).Elem().Interface()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Unmarshal(...) error got %v, want %v", got, want)
			}
			if got, want := value, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Unmarshal(...) value got %v, want %v", got, want)
			}
		})
	}
}

func TestUnmarshal_Panics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		target any
	}{
		{
			name:   "NonPointer",
			target: 42,
		}, {
			name:   "NilPointer",
			target: (*int)(nil),
		}, {
			name:   "UnsupportedKind",
			target: new(struct{ Field int }),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var recovered any

			// Act
			func() {
				defer func() { recovered = recover() }()
				_ = ask.Unmarshal("value", tc.target)
			}()

			// Assert
			message, ok := recovered.(string)
			if !ok || !strings.Contains(message, "ask:") {
				t.Fatalf("Unmarshal(...) panic got %v, want a message containing %q", recovered, "ask:")
			}
		})
	}
}
