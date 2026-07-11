package help_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/help"
	"github.com/bitwizeshift/go-cli/internal/template/help/helptest"
	"github.com/bitwizeshift/go-cli/internal/template/plain"
)

func TestRenderer_Render_Golden(t *testing.T) {
	t.Parallel()

	testCases := helptest.Cases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			want := readGolden(t, tc.Name)
			sut := help.Renderer{Columns: tc.Columns}
			var buf bytes.Buffer

			// Act
			renderErr := sut.Render(&buf, tc.Command)
			rendered, stripErr := plain.Render(buf.String())

			// Assert
			if renderErr != nil {
				t.Fatalf("Render(...) = %v, want nil", renderErr)
			}
			if stripErr != nil {
				t.Fatalf("plain.Render(...) = %v, want nil", stripErr)
			}
			if got, want := rendered, want; !cmp.Equal(got, want) {
				t.Errorf("Render(...) mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

// readGolden reads the golden file named name from the testdata directory.
func readGolden(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("readGolden(%q) = %v, want nil", name, err)
	}
	return string(data)
}
