package richtext_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/richtext"
	"github.com/bitwizeshift/go-cli/richtext/style"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	reset    = "\x1b[0m"
	red      = "\x1b[31m"
	blueBg   = "\x1b[44m"
	bold     = "\x1b[1m"
	boldItal = "\x1b[1;3m"
	title    = "\x1b[1;36m" // cyan + bold, from newTheme
)

func newTheme(t *testing.T) *richtext.Theme {
	t.Helper()

	theme, err := richtext.NewTheme(struct {
		Title    style.Style `theme:"title"`
		Note     style.Style // exported but untagged: ignored
		reserved style.Style `theme:"reserved"` // unexported: ignored
	}{
		Title: style.Style{Foreground: style.Cyan, Attributes: style.Bold},
	})
	if err != nil {
		t.Fatalf("NewTheme() = %v, want nil", err)
	}
	return theme
}

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	theme := newTheme(t)

	testCases := []struct {
		name    string
		theme   *richtext.Theme
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "Foreground",
			theme:   nil,
			input:   "[fg:red]x[/fg]",
			want:    reset + red + "x" + reset,
			wantErr: nil,
		},
		{
			name:    "Background",
			theme:   nil,
			input:   "[bg:blue]x[/bg]",
			want:    reset + blueBg + "x" + reset,
			wantErr: nil,
		},
		{
			name:    "AttributesAccumulateThenUnwind",
			theme:   nil,
			input:   "[attr:bold]a[attr:italic]b[/attr]c[/attr]",
			want:    reset + bold + "a" + reset + boldItal + "b" + reset + bold + "c" + reset,
			wantErr: nil,
		},
		{
			name:    "UnknownNamespacePassesThrough",
			theme:   nil,
			input:   "p[foo:bar]q[/foo]r",
			want:    "p[foo:bar]q[/foo]r",
			wantErr: nil,
		},
		{
			name:    "UnknownFieldResetsActiveStyle",
			theme:   nil,
			input:   "[fg:red][fg:bogus]x[/fg][/fg]",
			want:    reset + red + reset + "x" + reset + red + reset,
			wantErr: nil,
		},
		{
			name:    "UnknownAttributeResetsActiveStyle",
			theme:   nil,
			input:   "[attr:bold][attr:sparkle]x[/attr][/attr]",
			want:    reset + bold + reset + "x" + reset + bold + reset,
			wantErr: nil,
		},
		{
			name:    "ThemeAppliesFullStyle",
			theme:   theme,
			input:   "[theme:title]x[/theme]",
			want:    reset + title + "x" + reset,
			wantErr: nil,
		},
		{
			name:    "UnknownThemeWithNilRegistryResets",
			theme:   nil,
			input:   "[theme:none]x[/theme]",
			want:    "x",
			wantErr: nil,
		},
		{
			name:    "UnbalancedClose",
			theme:   nil,
			input:   "[fg:red]x[/bg]",
			want:    reset + red + "x",
			wantErr: richtext.ErrUnbalancedTag,
		},
		{
			name:    "RawRegionWritesContentsVerbatim",
			theme:   nil,
			input:   "[richtext:off][fg:red]x[/richtext]",
			want:    "[fg:red]x",
			wantErr: nil,
		},
		{
			name:    "RawRegionInsideThemeKeepsStyle",
			theme:   theme,
			input:   "[theme:title][richtext:off][x][/richtext][/theme]",
			want:    reset + title + "[x]" + reset,
			wantErr: nil,
		},
		{
			name:    "RichtextUnknownFieldResetsActiveStyle",
			theme:   nil,
			input:   "[fg:red][richtext:on]x[/richtext][/fg]",
			want:    reset + red + reset + "x" + reset + red + reset,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf strings.Builder
			sut := richtext.NewWriter(&buf, tc.theme)
			sut.ForceColour()

			// Act
			n, err := sut.Write([]byte(tc.input))

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Write() = %v, want %v", got, want)
			}
			if got, want := n, len(tc.input); !cmp.Equal(got, want) {
				t.Errorf("Write() n = %d, want %d", got, want)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Write() output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriter_EnableColour(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		enable bool
		input  string
		want   string
	}{
		{
			name:   "DisabledSuppressesColour",
			enable: false,
			input:  "[fg:red]x[/fg]",
			want:   "x",
		},
		{
			name:   "DisabledSuppressesAttributes",
			enable: false,
			input:  "[attr:bold]a[attr:italic]b[/attr]c[/attr]",
			want:   "abc",
		},
		{
			name:   "DisabledPassesUnknownNamespace",
			enable: false,
			input:  "p[foo:bar]q[/foo]r",
			want:   "p[foo:bar]q[/foo]r",
		},
		{
			name:   "EnabledOnNonTTYSuppressesColour",
			enable: true,
			input:  "[fg:red]x[/fg]",
			want:   "x",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf strings.Builder
			sut := richtext.NewWriter(&buf, nil)
			sut.EnableColour(tc.enable)

			// Act
			_, err := sut.Write([]byte(tc.input))

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Write() = %v, want nil", got)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Write() output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriter_Close(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "Balanced",
			input:   "[fg:red]x[/fg]",
			want:    reset + red + "x" + reset,
			wantErr: nil,
		},
		{
			name:    "UnclosedTag",
			input:   "[fg:red]x",
			want:    reset + red + "x",
			wantErr: richtext.ErrUnclosedTag,
		},
		{
			name:    "TrailingPartialTagFlushedAsText",
			input:   "ab[fg",
			want:    "ab[fg",
			wantErr: nil,
		},
		{
			name:    "UnclosedRawRegion",
			input:   "[richtext:off]abc",
			want:    "abc",
			wantErr: richtext.ErrUnclosedTag,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var buf strings.Builder
			sut := richtext.NewWriter(&buf, nil)
			sut.ForceColour()
			_, _ = sut.Write([]byte(tc.input))

			// Act
			err := sut.Close()

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Close() = %v, want %v", got, want)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Close() output = %q, want %q", got, want)
			}
		})
	}
}

func TestWriter_Writer(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf strings.Builder
	sut := richtext.NewWriter(&buf, nil)

	// Act
	underlying := sut.Writer()

	// Assert
	sameWriter := cmp.Comparer(func(a, b io.Writer) bool { return a == b })
	if got, want := underlying, io.Writer(&buf); !cmp.Equal(got, want, sameWriter) {
		t.Errorf("Writer() = %v, want %v", got, want)
	}
}

func TestWriter_Write_UnbalancedClose_ReturnsTagError(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf strings.Builder
	sut := richtext.NewWriter(&buf, nil)
	input := []byte("[fg:red][/bg]")

	// Act
	_, err := sut.Write(input)

	// Assert
	var tagErr *richtext.TagError
	if !errors.As(err, &tagErr) {
		t.Fatalf("Write() = %v, want *richtext.TagError", err)
	}
	if got, want := tagErr.Namespace, "bg"; !cmp.Equal(got, want) {
		t.Errorf("Namespace = %q, want %q", got, want)
	}
}

// errWriter fails once it has accepted okWrites successful writes.
type errWriter struct {
	okWrites int
	writes   int
	err      error
}

func (w *errWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes > w.okWrites {
		return 0, w.err
	}
	return len(p), nil
}

func TestWriter_Close_FlushWriteFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	wantErr := errors.New("boom")
	dst := &errWriter{okWrites: 1, err: wantErr}
	sut := richtext.NewWriter(dst, nil)
	_, _ = sut.Write([]byte("ab[fg"))

	// Act
	err := sut.Close()

	// Assert
	if got, want := err, wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Close() = %v, want %v", got, want)
	}
}
