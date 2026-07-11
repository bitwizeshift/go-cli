package version_test

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/plain"
	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
	"github.com/bitwizeshift/go-cli/internal/template/version/versiontest"
)

func TestTemplate_Golden(t *testing.T) {
	// Arrange
	restore := tmplfuncs.DefaultBuild.ReadBuildInfo
	t.Cleanup(func() { tmplfuncs.DefaultBuild.ReadBuildInfo = restore })
	tmplfuncs.DefaultBuild.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
		return versiontest.BuildInfo(), true
	}
	want := readGolden(t, "version.golden.txt")
	cmd := versiontest.Command()

	// Act
	markup, renderErr := versiontest.Render(cmd)
	rendered, stripErr := plain.Render(markup)

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
