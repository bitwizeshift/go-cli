package term_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/term"
)

func TestIsTTYFuncEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		hasFd    bool
		fnResult bool
		want     bool
	}{
		{
			name:     "FdWriterFuncReturnsTrue",
			hasFd:    true,
			fnResult: true,
			want:     true,
		}, {
			name:     "FdWriterFuncReturnsFalse",
			hasFd:    true,
			fnResult: false,
			want:     false,
		}, {
			name:     "WriterWithoutFd",
			hasFd:    false,
			fnResult: true,
			want:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var writer io.Writer
			if tc.hasFd {
				writer = &fdWriter{fd: 1}
			} else {
				writer = &bytes.Buffer{}
			}
			enabler := term.IsTTYFuncEnabler(func(int) bool { return tc.fnResult })

			// Act
			got := enabler.EnableColour(writer)

			// Assert
			if got, want := got, tc.want; !cmp.Equal(got, want) {
				t.Errorf("IsTTYFuncEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestFixedEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		result bool
		want   bool
	}{
		{
			name:   "True",
			result: true,
			want:   true,
		}, {
			name:   "False",
			result: false,
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := term.FixedEnabler(tc.result)

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FixedEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestEnvEnabler_EnableColour(t *testing.T) {
	const varName = "ANSI_TEST_ENV_ENABLER_VAR"

	testCases := []struct {
		name  string
		set   bool
		value string
		want  bool
	}{
		{
			name:  "NotSet",
			set:   false,
			value: "",
			want:  false,
		}, {
			name:  "SetTrue",
			set:   true,
			value: "true",
			want:  true,
		}, {
			name:  "SetOne",
			set:   true,
			value: "1",
			want:  true,
		}, {
			name:  "SetFalse",
			set:   true,
			value: "false",
			want:  false,
		}, {
			name:  "SetZero",
			set:   true,
			value: "0",
			want:  false,
		}, {
			name:  "SetEmpty",
			set:   true,
			value: "",
			want:  false,
		}, {
			name:  "SetGarbage",
			set:   true,
			value: "garbage",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			if tc.set {
				t.Setenv(varName, tc.value)
			}
			sut := term.EnvEnabler{Variable: varName}

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("EnvEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestInvertEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		innerVal bool
		want     bool
	}{
		{
			name:     "InnerTrue",
			innerVal: true,
			want:     false,
		}, {
			name:     "InnerFalse",
			innerVal: false,
			want:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := term.InvertEnabler{
				Enabler: term.FixedEnabler(tc.innerVal),
			}

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("InvertEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}

func TestConjunctiveEnabler_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		enablers []term.ColourEnabler
		want     bool
	}{
		{
			name:     "Empty",
			enablers: nil,
			want:     false,
		}, {
			name:     "SingleTrue",
			enablers: []term.ColourEnabler{term.FixedEnabler(true)},
			want:     true,
		}, {
			name:     "SingleFalse",
			enablers: []term.ColourEnabler{term.FixedEnabler(false)},
			want:     false,
		}, {
			name: "MultipleAllTrue",
			enablers: []term.ColourEnabler{
				term.FixedEnabler(true),
				term.FixedEnabler(true),
				term.FixedEnabler(true),
			},
			want: true,
		}, {
			name: "MultipleFirstFalse",
			enablers: []term.ColourEnabler{
				term.FixedEnabler(false),
				term.FixedEnabler(true),
				term.FixedEnabler(true),
			},
			want: false,
		}, {
			name: "MultipleLastFalse",
			enablers: []term.ColourEnabler{
				term.FixedEnabler(true),
				term.FixedEnabler(true),
				term.FixedEnabler(false),
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := term.ConjunctiveEnabler(tc.enablers)

			// Act
			enabled := sut.EnableColour(&bytes.Buffer{})

			// Assert
			if got, want := enabled, tc.want; !cmp.Equal(got, want) {
				t.Errorf("ConjunctiveEnabler.EnableColour(...) got %v, want %v", got, want)
			}
		})
	}
}
