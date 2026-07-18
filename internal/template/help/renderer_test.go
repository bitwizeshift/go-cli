package help_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
)

func TestRenderer_Render_Notice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		notice *help.Notice
		want   string
	}{
		{
			name:   "WithNotice",
			notice: &help.Notice{Current: "v1.0.0", Latest: "v2.0.0"},
			want:   "A new version is available: v1.0.0 → v2.0.0",
		}, {
			name:   "WithoutNotice",
			notice: nil,
			want:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := help.Renderer{Columns: 80, Notice: tc.notice}
			command := &cobra.Command{Use: "app"}
			var buf bytes.Buffer

			// Act
			renderErr := sut.Render(&buf, command, nil)
			rendered, stripErr := plain.Render(buf.String())

			// Assert
			if got, want := renderErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Renderer.Render() = %v, want %v", got, want)
			}
			if got, want := stripErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("plain.Render() = %v, want %v", got, want)
			}
			if got, want := noticeLine(rendered), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Renderer.Render() advisory = %q, want %q", got, want)
			}
		})
	}
}

// noticeLine returns the trimmed advisory line from rendered help output, or the
// empty string when no advisory is present.
func noticeLine(rendered string) string {
	for line := range strings.SplitSeq(rendered, "\n") {
		if strings.Contains(line, "A new version is available:") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
