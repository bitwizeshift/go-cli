package update

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// ensureV returns version with a leading "v", which [semver] requires. An empty
// string and an already-prefixed version are returned unchanged.
func ensureV(version string) string {
	if version == "" || strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

// canonicalVersion normalizes a version reported by a channel into canonical,
// v-prefixed semver. It returns [ErrInvalidVersion] when version is not valid
// semver.
func canonicalVersion(version string) (string, error) {
	v := ensureV(version)
	if !semver.IsValid(v) {
		return "", fmt.Errorf("%w: %q", ErrInvalidVersion, version)
	}
	return semver.Canonical(v), nil
}

// isNewer reports whether latest is a strictly greater semantic version than
// current. A version that is not valid semver yields false, so a non-release
// current version such as "snapshot" never reports an update.
func isNewer(latest, current string) bool {
	latest, current = ensureV(latest), ensureV(current)
	if !semver.IsValid(latest) || !semver.IsValid(current) {
		return false
	}
	return semver.Compare(latest, current) > 0
}
