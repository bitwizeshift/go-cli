package tmplfuncs

import (
	"runtime"
	"runtime/debug"
)

// Build exposes build and version information to templates. Its methods read
// from [debug.ReadBuildInfo] via the injectable [Build.ReadBuildInfo] hook so
// the values can be controlled in tests.
type Build struct {
	// ReadBuildInfo reports the build information of the running binary. It
	// defaults to [debug.ReadBuildInfo] and may be replaced in tests.
	ReadBuildInfo func() (info *debug.BuildInfo, ok bool)
}

// DefaultBuild reads the running binary's build information via
// [debug.ReadBuildInfo].
var DefaultBuild = Build{
	ReadBuildInfo: debug.ReadBuildInfo,
}

// Field is a single labelled build-detail entry.
type Field struct {
	Name  string
	Value string
}

// Target returns the compilation target as "GOOS/GOARCH", using the build
// settings when present and falling back to the running binary's values.
func (b *Build) Target() string {
	return b.settingOr("GOOS", runtime.GOOS) + "/" + b.settingOr("GOARCH", runtime.GOARCH)
}

// Tags returns the build tags the binary was compiled with, or the empty string
// when none were set.
func (b *Build) Tags() string {
	return b.setting("-tags")
}

// GoVersion returns the Go toolchain version that built the binary, falling back
// to the running toolchain's version.
func (b *Build) GoVersion() string {
	if info, ok := b.ReadBuildInfo(); ok && info.GoVersion != "" {
		return info.GoVersion
	}
	return runtime.Version()
}

// VCS returns the version-control system that built the binary, or "unknown".
func (b *Build) VCS() string {
	return b.settingOr("vcs", "unknown")
}

// VCSRevision returns the version-control revision of the build, or "unknown".
func (b *Build) VCSRevision() string {
	return b.settingOr("vcs.revision", "unknown")
}

// VCSTime returns the version-control commit time of the build, or "unknown".
func (b *Build) VCSTime() string {
	return b.settingOr("vcs.time", "unknown")
}

// Fields returns the ordered build-detail rows suitable for a labelled listing.
func (b *Build) Fields() []Field {
	return []Field{
		{Name: "Target", Value: b.Target()},
		{Name: "Build Tags", Value: b.Tags()},
		{Name: "Go Version", Value: b.GoVersion()},
		{Name: "VCS", Value: b.VCS()},
		{Name: "VCS Revision", Value: b.VCSRevision()},
		{Name: "VCS Time", Value: b.VCSTime()},
	}
}

// FieldNames returns the labels of [Build.Fields] in order.
func (b *Build) FieldNames() []string {
	fields := b.Fields()
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.Name
	}
	return names
}

// settingOr returns the build setting for key, or fallback when it is absent.
func (b *Build) settingOr(key, fallback string) string {
	if v := b.setting(key); v != "" {
		return v
	}
	return fallback
}

// setting returns the value of the build setting for key, or the empty string
// when it is absent.
func (b *Build) setting(key string) string {
	for _, setting := range b.settings() {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}

// settings returns the build settings of the running binary, or nil when the
// build information is unavailable.
func (b *Build) settings() []debug.BuildSetting {
	if info, ok := b.ReadBuildInfo(); ok {
		return info.Settings
	}
	return nil
}
