package tmplfuncs_test

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/template/tmplfuncs"
)

// readerOf returns a ReadBuildInfo hook yielding info and ok.
func readerOf(info *debug.BuildInfo, ok bool) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) { return info, ok }
}

func TestBuild_GoVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{
			name: "reports embedded toolchain",
			info: &debug.BuildInfo{GoVersion: "go1.24.0"},
			ok:   true,
			want: "go1.24.0",
		}, {
			name: "falls back to running toolchain",
			info: nil,
			ok:   false,
			want: runtime.Version(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(tc.info, tc.ok)}

			// Act
			goVersion := sut.GoVersion()

			// Assert
			if got, want := goVersion, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Build.GoVersion() = %q, want %q", got, want)
			}
		})
	}
}

func TestBuild_Target(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name: "from build settings",
			settings: []debug.BuildSetting{
				{Key: "GOOS", Value: "linux"},
				{Key: "GOARCH", Value: "amd64"},
			},
			want: "linux/amd64",
		}, {
			name:     "falls back to running target",
			settings: nil,
			want:     runtime.GOOS + "/" + runtime.GOARCH,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			info := &debug.BuildInfo{Settings: tc.settings}
			sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(info, true)}

			// Act
			target := sut.Target()

			// Assert
			if got, want := target, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Build.Target() = %q, want %q", got, want)
			}
		})
	}
}

func TestBuild_Tags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name:     "reports build tags",
			settings: []debug.BuildSetting{{Key: "-tags", Value: "netgo,osusergo"}},
			want:     "netgo,osusergo",
		}, {
			name:     "empty when unset",
			settings: nil,
			want:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			info := &debug.BuildInfo{Settings: tc.settings}
			sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(info, true)}

			// Act
			tags := sut.Tags()

			// Assert
			if got, want := tags, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Build.Tags() = %q, want %q", got, want)
			}
		})
	}
}

func TestBuildVCS(t *testing.T) {
	t.Parallel()

	settings := []debug.BuildSetting{
		{Key: "vcs", Value: "git"},
		{Key: "vcs.revision", Value: "abc1234"},
		{Key: "vcs.time", Value: "2024-08-05T20:18:23Z"},
	}

	testCases := []struct {
		name     string
		settings []debug.BuildSetting
		method   func(*tmplfuncs.Build) string
		want     string
	}{
		{
			name:     "vcs present",
			settings: settings,
			method:   (*tmplfuncs.Build).VCS,
			want:     "git",
		}, {
			name:     "revision present",
			settings: settings,
			method:   (*tmplfuncs.Build).VCSRevision,
			want:     "abc1234",
		}, {
			name:     "time present",
			settings: settings,
			method:   (*tmplfuncs.Build).VCSTime,
			want:     "2024-08-05T20:18:23Z",
		}, {
			name:     "absent reports unknown",
			settings: nil,
			method:   (*tmplfuncs.Build).VCS,
			want:     "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			info := &debug.BuildInfo{Settings: tc.settings}
			sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(info, true)}

			// Act
			value := tc.method(sut)

			// Assert
			if got, want := value, tc.want; !cmp.Equal(got, want) {
				t.Errorf("method(Build) = %q, want %q", got, want)
			}
		})
	}
}

func TestBuild_Fields(t *testing.T) {
	t.Parallel()

	// Arrange
	info := &debug.BuildInfo{
		GoVersion: "go1.24.0",
		Settings: []debug.BuildSetting{
			{Key: "GOOS", Value: "linux"},
			{Key: "GOARCH", Value: "amd64"},
			{Key: "-tags", Value: "netgo"},
			{Key: "vcs", Value: "git"},
			{Key: "vcs.revision", Value: "abc1234"},
			{Key: "vcs.time", Value: "2024-08-05T20:18:23Z"},
		},
	}
	sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(info, true)}
	want := []tmplfuncs.Field{
		{Name: "Target", Value: "linux/amd64"},
		{Name: "Build Tags", Value: "netgo"},
		{Name: "Go Version", Value: "go1.24.0"},
		{Name: "VCS", Value: "git"},
		{Name: "VCS Revision", Value: "abc1234"},
		{Name: "VCS Time", Value: "2024-08-05T20:18:23Z"},
	}

	// Act
	fields := sut.Fields()

	// Assert
	if got, want := fields, want; !cmp.Equal(got, want) {
		t.Errorf("Build.Fields() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func TestBuild_FieldNames(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := &tmplfuncs.Build{ReadBuildInfo: readerOf(&debug.BuildInfo{}, true)}
	want := []string{"Target", "Build Tags", "Go Version", "VCS", "VCS Revision", "VCS Time"}

	// Act
	names := sut.FieldNames()

	// Assert
	if got, want := names, want; !cmp.Equal(got, want) {
		t.Errorf("Build.FieldNames() mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}
