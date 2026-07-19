package buildinfo

import "runtime/debug"

// SnapshotVersion is the version of a binary built without an embedded module
// version, such as one built straight from a working tree.
const SnapshotVersion = "snapshot"

// VersionReader reports the module version embedded in the running binary.
type VersionReader struct {
	// ReadBuildInfo reports the build information of the running binary. It
	// defaults to [debug.ReadBuildInfo] and may be replaced in tests.
	ReadBuildInfo func() (info *debug.BuildInfo, ok bool)
}

// DefaultVersionReader reads the running binary's build information via
// [debug.ReadBuildInfo].
var DefaultVersionReader = VersionReader{
	ReadBuildInfo: debug.ReadBuildInfo,
}

// Version returns the module version of the running binary, or
// [SnapshotVersion] when it carries none.
func (vr *VersionReader) Version() string {
	if info, ok := vr.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return SnapshotVersion
}
