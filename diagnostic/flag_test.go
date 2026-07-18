package diagnostic_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/clitest"
	"github.com/bitwizeshift/go-cli/diagnostic"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// parseFormat registers lf into a fresh registry and parses args, failing the
// test on a parse error.
func parseFormat(t *testing.T, lf *diagnostic.LoggerFlag, args ...string) {
	t.Helper()

	cl := argtest.NewCommandLine()
	lf.RegisterArgs(cl)
	argtest.Parse(t, cl, args...)
}

func TestFormatType_UnmarshalText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    diagnostic.FormatType
		wantErr error
	}{
		{
			name:    "Text",
			input:   "text",
			want:    diagnostic.FormatText,
			wantErr: nil,
		}, {
			name:    "GitHub",
			input:   "github",
			want:    diagnostic.FormatGitHub,
			wantErr: nil,
		}, {
			name:    "JSON",
			input:   "json",
			want:    diagnostic.FormatJSON,
			wantErr: nil,
		}, {
			name:    "Unknown",
			input:   "xml",
			want:    diagnostic.FormatType(""),
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var format diagnostic.FormatType

			// Act
			err := format.UnmarshalText([]byte(tc.input))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("FormatType.UnmarshalText(%q) = %v, want %v", tc.input, got, want)
			}
			if got, want := format, tc.want; !cmp.Equal(got, want) {
				t.Errorf("FormatType.UnmarshalText(%q) format = %q, want %q", tc.input, got, want)
			}
		})
	}
}

func TestLoggerFlag_RegisterArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		flag diagnostic.LoggerFlag
		want []*argtest.Flag
	}{
		{
			name: "Defaults",
			flag: diagnostic.LoggerFlag{},
			want: []*argtest.Flag{
				{Long: "output-format", Type: "format"},
			},
		}, {
			name: "CustomLongFlag",
			flag: diagnostic.LoggerFlag{LongFlag: "fmt"},
			want: []*argtest.Flag{
				{Long: "fmt", Type: "format"},
			},
		}, {
			name: "CustomShortFlag",
			flag: diagnostic.LoggerFlag{LongFlag: "fmt", ShortFlag: "f"},
			want: []*argtest.Flag{
				{Long: "fmt", Short: "f", Type: "format"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			lf := tc.flag

			// Act
			lf.RegisterArgs(cl)
			flags := argtest.AllFlags(cl)

			// Assert
			if got, want := flags, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("RegisterArgs(...) flags = %+v, want %+v", got, want)
			}
		})
	}
}

func TestLoggerFlag_LoggerFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "DefaultFormat",
			args: nil,
		}, {
			name: "Text",
			args: []string{"--output-format", "text"},
		}, {
			name: "GitHub",
			args: []string{"--output-format", "github"},
		}, {
			name: "JSON",
			args: []string{"--output-format", "json"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			sut := &diagnostic.LoggerFlag{}
			parseFormat(t, sut, tc.args...)

			// Act
			logger := sut.LoggerFor(&buf)
			logger.Error(context.Background(), &diagnostic.Diagnostic{
				Message: "test",
			})

			// Assert
			if got, want := buf.String() != "", true; !cmp.Equal(got, want) {
				t.Errorf("LoggerFor(...).Error(...) = non-empty %v, want %v", got, want)
			}
		})
	}
}

func TestLoggerFlag_Logger(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "DefaultFormat",
			args: nil,
		}, {
			name: "Text",
			args: []string{"--output-format", "text"},
		}, {
			name: "GitHub",
			args: []string{"--output-format", "github"},
		}, {
			name: "JSON",
			args: []string{"--output-format", "json"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf bytes.Buffer
			ctx := clitest.WithWriters(context.Background(), &bytes.Buffer{}, &buf)
			sut := &diagnostic.LoggerFlag{}
			parseFormat(t, sut, tc.args...)

			// Act
			logger := sut.Logger(ctx)
			logger.Error(context.Background(), &diagnostic.Diagnostic{
				Message: "test",
			})

			// Assert
			if got, want := buf.String() != "", true; !cmp.Equal(got, want) {
				t.Errorf("LoggerFor(...).Error(...) = non-empty %v, want %v", got, want)
			}
		})
	}
}
