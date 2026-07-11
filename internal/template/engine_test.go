package template_test

import (
	"bytes"
	"errors"
	"io"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/bitwizeshift/go-cli/internal/template/panichandler"
	"github.com/bitwizeshift/go-cli/internal/template/version"
	"github.com/bitwizeshift/go-cli/internal/term"
	"github.com/bitwizeshift/go-cli/richtext"
)

// capturingSizer records the writer it is asked to size and reports a fixed
// width, so tests can observe which writer the engine measures.
type capturingSizer struct {
	got     io.Writer
	columns int
}

func (s *capturingSizer) Columns(w io.Writer) int {
	s.got = w
	return s.columns
}

func TestRenderEngine_HelpRenderer_SizesUnwrappedWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	var base bytes.Buffer
	wrapped := richtext.NewWriter(&base, nil)
	sizer := &capturingSizer{columns: 72}
	sut := template.RenderEngine{Sizer: sizer}

	// Act
	renderer := sut.HelpRenderer(wrapped)

	// Assert
	sameWriter := cmp.Comparer(func(a, b io.Writer) bool { return a == b })
	if got, want := sizer.got, io.Writer(&base); !cmp.Equal(got, want, sameWriter) {
		t.Errorf("HelpRenderer sized %v, want the unwrapped %v", got, want)
	}
	if got, want := renderer.Columns, 72; !cmp.Equal(got, want) {
		t.Errorf("HelpRenderer(...).Columns = %d, want %d", got, want)
	}
}

func TestRenderEngine_HelpFunc_WritesStyledHelp(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "app"}
	cmd.SetOut(&buf)

	// Act
	sut.HelpFunc()(cmd, nil)

	// Assert
	if got, want := strings.Contains(buf.String(), "[theme:title]app[/theme]"), true; !cmp.Equal(got, want) {
		t.Errorf("HelpFunc() output = %q, want it to contain the styled title", buf.String())
	}
}

func TestRenderEngine_UsageFunc_WritesStyledUsage(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "app"}
	cmd.SetErr(&buf)

	// Act
	err := sut.UsageFunc()(cmd)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("UsageFunc()(cmd) = %v, want nil", err)
	}
	if got, want := strings.Contains(buf.String(), "[theme:title]app[/theme]"), true; !cmp.Equal(got, want) {
		t.Errorf("UsageFunc() output = %q, want it to contain the styled title", buf.String())
	}
}

func TestRenderEngine_Errorf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		columns int
		format  string
		args    []any
		want    string
	}{
		{
			name:    "SingleLine",
			columns: 80,
			format:  "boom %d",
			args:    []any{42},
			want:    "[theme:error]error: [/theme][richtext:off]boom 42[/richtext]",
		},
		{
			name:    "WrapsToContinuationLines",
			columns: 20,
			format:  "one two three four",
			args:    nil,
			want: "[theme:error]error: [/theme][richtext:off]one two three[/richtext]\n" +
				"       [richtext:off]four[/richtext]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{Sizer: term.FixedSizer(tc.columns)}
			var buf bytes.Buffer

			// Act
			err := sut.Errorf(&buf, tc.format, tc.args...)

			// Assert
			if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Errorf(...) = %v, want nil", err)
			}
			if got, want := buf.String(), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Errorf(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestRenderEngine_PanicRenderer_WritesStyledReport(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}
	ctx := panichandler.PanicContext{
		Err:   errors.New("boom"),
		Stack: []byte("goroutine 1 [running]:\nmain.main()"),
	}
	var buf bytes.Buffer

	// Act
	err := sut.PanicRenderer().Render(&buf, ctx)

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Render(...) = %v, want nil", err)
	}
	if got, want := strings.Contains(buf.String(), "[theme:error]error:[/theme]"), true; !cmp.Equal(got, want) {
		t.Errorf("PanicRenderer().Render() output = %q, want it to contain the styled error label", buf.String())
	}
}

func TestRenderEngine_FlagErrorFunc_StylesError(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}
	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "app"}
	cmd.SetErr(&buf)

	// Act
	err := sut.FlagErrorFunc()(cmd, errors.New("unknown flag"))

	// Assert
	if got, want := err, (error)(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("FlagErrorFunc()(cmd, err) = %v, want nil", err)
	}
	want := "[theme:error]error: [/theme][richtext:off]unknown flag[/richtext]"
	if got, want := buf.String(), want; !cmp.Equal(got, want) {
		t.Errorf("FlagErrorFunc()(cmd, err) output = %q, want %q", got, want)
	}
}

func TestRenderEngine_VersionTemplate(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}

	// Act
	tmpl := sut.VersionTemplate()

	// Assert
	if got, want := tmpl, version.Template(); !cmp.Equal(got, want) {
		t.Errorf("VersionTemplate() = %q, want %q", got, want)
	}
}

func TestRenderEngine_VersionFuncs_Keys(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{Sizer: term.FixedSizer(80)}
	modules := []string{"build", "text"}

	// Act
	funcs := sut.VersionFuncs()
	keys := slices.Collect(maps.Keys(funcs))

	// Assert
	opts := cmp.Options{cmpopts.SortSlices(strings.Compare)}
	if got, want := keys, modules; !cmp.Equal(got, want, opts...) {
		t.Errorf("VersionFuncs() = mismatch (-want +got):\n%s", cmp.Diff(got, want, opts...))
	}
}
