package buildinfo_test

import (
	"runtime/debug"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/buildinfo"
)

// readerOf returns a ReadBuildInfo hook yielding info and ok.
func readerOf(info *debug.BuildInfo, ok bool) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) { return info, ok }
}

func TestVersionReader_Version(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "ReleasedVersion",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			ok:   true,
			want: "v1.2.3",
		}, {
			name: "DevelVersion",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: buildinfo.SnapshotVersion,
		}, {
			name: "EmptyVersion",
			info: &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:   true,
			want: buildinfo.SnapshotVersion,
		}, {
			name: "BuildInfoUnavailable",
			info: nil,
			ok:   false,
			want: buildinfo.SnapshotVersion,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &buildinfo.VersionReader{ReadBuildInfo: readerOf(tc.info, tc.ok)}

			// Act
			version := sut.Version()

			// Assert
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("VersionReader.Version() = %q, want %q", got, want)
			}
		})
	}
}
