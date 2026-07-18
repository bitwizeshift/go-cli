package arg_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/internal/argreg"
)

// firstPositional returns the [argreg.Positional] registered on cl at position
// zero, so a test can drive its decode closure directly.
func firstPositional(cl *arg.CommandLine) *argreg.Positional {
	return argreg.Positionals((*argreg.CommandLine)(cl))[0]
}

func TestPositional_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		value   string
		want    string
		wantErr error
	}{
		{
			name:    "DecodesValue",
			options: nil,
			value:   "alpha",
			want:    "alpha",
			wantErr: nil,
		}, {
			name:    "TransformsValueWithDecoder",
			options: []arg.Option{arg.UnmarshalWith(yell)},
			value:   "abc",
			want:    "ABC",
			wantErr: nil,
		}, {
			name:    "DecoderErrorPropagates",
			options: []arg.Option{arg.UnmarshalWith(failString)},
			value:   "x",
			want:    "",
			wantErr: errDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst string
			arg.Positional(cl, "value", 0, &dst, tc.options...)
			positional := firstPositional(cl)

			// Act
			err := positional.Set(tc.value)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Positional.Set(...) value = %q, want %q", got, want)
			}
		})
	}
}

func TestPositional_SetDecodesTypedValue(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst int
	arg.Positional(cl, "count", 0, &dst)
	positional := firstPositional(cl)

	// Act
	err := positional.Set("42")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
	if got, want := dst, 42; !cmp.Equal(got, want) {
		t.Errorf("Positional.Set(...) value = %d, want %d", got, want)
	}
}

func TestPositional_SetInvokesCallback(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst, seen string
	arg.Positional(cl, "value", 0, &dst, arg.Callback(func(s string) { seen = s }))
	positional := firstPositional(cl)

	// Act
	err := positional.Set("hello")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
	if got, want := seen, "hello"; !cmp.Equal(got, want) {
		t.Errorf("Positional.Set(...) callback value = %q, want %q", got, want)
	}
}

func TestPositional_SetCallbackError(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst string
	arg.Positional(cl, "value", 0, &dst, arg.Callback(func(string) error { return errDecode }))
	positional := firstPositional(cl)

	// Act
	err := positional.Set("x")

	// Assert
	if got, want := err, errDecode; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
}

func TestPositional_Metadata(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var name string
	var count int
	arg.Positional(cl, "name", 0, &name, arg.Usage("the name"))
	arg.Positional(cl, "count", 1, &count, arg.Type("number"), arg.Usage("how many"))

	// Act
	positionals := argtest.AllPositionals(cl)

	// Assert
	want := []*argtest.Positional{
		{Index: 0, Name: "name", Type: "string", Usage: "the name"},
		{Index: 1, Name: "count", Type: "number", Usage: "how many"},
	}
	if got, want := positionals, want; !cmp.Equal(got, want) {
		t.Errorf("AllPositionals(...) = %+v, want %+v", got, want)
	}
}

func TestUnmatched_AssignsValues(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var rest []string
	arg.Unmatched(cl, &rest)
	unmatched := argreg.GetUnmatched((*argreg.CommandLine)(cl))

	// Act
	err := unmatched.Set([]string{"a", "b"})

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
	}
	if got, want := rest, []string{"a", "b"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Unmatched.Set(...) values = %v, want %v", got, want)
	}
}
