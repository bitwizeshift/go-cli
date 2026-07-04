package template_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"

	"github.com/bitwizeshift/go-cli/internal/ansi"
	"github.com/bitwizeshift/go-cli/internal/template"
	"github.com/bitwizeshift/go-cli/internal/template/palette"
	"github.com/bitwizeshift/go-cli/internal/template/version"
	"github.com/bitwizeshift/go-cli/internal/term"
)

// containsANSI reports whether s contains an ANSI escape sequence.
func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b")
}

func TestRenderEngine_HelpRenderer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		enabled     bool
		columns     int
		wantPalette palette.Palette
	}{
		{
			name:        "colour enabled",
			enabled:     true,
			columns:     80,
			wantPalette: palette.DefaultColour,
		}, {
			name:        "colour disabled",
			enabled:     false,
			columns:     60,
			wantPalette: palette.NoColour{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(tc.columns),
			}

			// Act
			renderer := sut.HelpRenderer(io.Discard)

			// Assert
			if got, want := renderer.Palette, tc.wantPalette; !cmp.Equal(got, want) {
				t.Errorf("HelpRenderer(...).Palette = %#v, want %#v", got, want)
			}
			if got, want := renderer.Columns, tc.columns; !cmp.Equal(got, want) {
				t.Errorf("HelpRenderer(...).Columns = %d, want %d", got, want)
			}
		})
	}
}

func TestRenderEngine_UsageRenderer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		enabled     bool
		wantPalette palette.Palette
	}{
		{
			name:        "colour enabled",
			enabled:     true,
			wantPalette: palette.DefaultColour,
		}, {
			name:        "colour disabled",
			enabled:     false,
			wantPalette: palette.NoColour{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(80),
			}

			// Act
			renderer := sut.UsageRenderer(io.Discard)

			// Assert
			if got, want := renderer.Palette, tc.wantPalette; !cmp.Equal(got, want) {
				t.Errorf("UsageRenderer(...).Palette = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRenderEngine_PanicRenderer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		enabled     bool
		wantPalette palette.Palette
	}{
		{
			name:        "colour enabled",
			enabled:     true,
			wantPalette: palette.DefaultColour,
		}, {
			name:        "colour disabled",
			enabled:     false,
			wantPalette: palette.NoColour{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(80),
			}

			// Act
			renderer := sut.PanicRenderer(io.Discard)

			// Assert
			if got, want := renderer.Palette, tc.wantPalette; !cmp.Equal(got, want) {
				t.Errorf("PanicRenderer(...).Palette = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRenderEngine_VersionFuncs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		enabled     bool
		wantPalette palette.Palette
	}{
		{
			name:        "colour enabled",
			enabled:     true,
			wantPalette: palette.DefaultColour,
		}, {
			name:        "colour disabled",
			enabled:     false,
			wantPalette: palette.NoColour{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(80),
			}
			funcs := sut.VersionFuncs()
			provider := funcs["palette"].(func() palette.Palette)

			// Act
			selected := provider()

			// Assert
			if got, want := selected, tc.wantPalette; !cmp.Equal(got, want) {
				t.Errorf("VersionFuncs()[palette]() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRenderEngine_VersionTemplate(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := template.RenderEngine{
		ColourEnabler: ansi.FixedEnabler(true),
		Sizer:         term.FixedSizer(80),
	}

	// Act
	tmpl := sut.VersionTemplate()

	// Assert
	if got, want := tmpl, version.Template(); !cmp.Equal(got, want) {
		t.Errorf("VersionTemplate() = %q, want %q", got, want)
	}
}

func TestRenderEngine_HelpFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		enabled      bool
		wantColoured bool
	}{
		{
			name:         "colour enabled",
			enabled:      true,
			wantColoured: true,
		}, {
			name:         "colour disabled",
			enabled:      false,
			wantColoured: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(80),
			}
			var buf bytes.Buffer
			cmd := &cobra.Command{Use: "app"}
			cmd.SetOut(&buf)

			// Act
			sut.HelpFunc()(cmd, nil)

			// Assert
			if got, want := containsANSI(buf.String()), tc.wantColoured; !cmp.Equal(got, want) {
				t.Errorf("HelpFunc()(cmd) coloured = %t, want %t", got, want)
			}
		})
	}
}

func TestRenderEngine_UsageFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		enabled      bool
		wantColoured bool
	}{
		{
			name:         "colour enabled",
			enabled:      true,
			wantColoured: true,
		}, {
			name:         "colour disabled",
			enabled:      false,
			wantColoured: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := template.RenderEngine{
				ColourEnabler: ansi.FixedEnabler(tc.enabled),
				Sizer:         term.FixedSizer(80),
			}
			var buf bytes.Buffer
			cmd := &cobra.Command{Use: "app"}
			cmd.SetErr(&buf)

			// Act
			err := sut.UsageFunc()(cmd)

			// Assert
			if err != nil {
				t.Fatalf("UsageFunc()(cmd) = %v, want nil", err)
			}
			if got, want := containsANSI(buf.String()), tc.wantColoured; !cmp.Equal(got, want) {
				t.Errorf("UsageFunc()(cmd) coloured = %t, want %t", got, want)
			}
		})
	}
}
