package updatecheck

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

var (
	// ErrInvalidVersion indicates a provider received a version that is not valid
	// semver.
	ErrInvalidVersion = errors.New("update: invalid version")
)

// ensureV returns version with a leading "v", which [semver] requires. An empty
// string and an already-prefixed version are returned unchanged.
func ensureV(version string) string {
	if version == "" || strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

// CanonicalVersion normalizes a version reported by a channel into canonical,
// v-prefixed semver. It returns [ErrInvalidVersion] when version is not valid
// semver.
func CanonicalVersion(version string) (string, error) {
	v := ensureV(version)
	if !semver.IsValid(v) {
		return "", fmt.Errorf("%w: %q", ErrInvalidVersion, version)
	}
	return semver.Canonical(v), nil
}

// IsNewer reports whether latest is a strictly greater semantic version than
// current. A version that is not valid semver yields false, so a non-release
// current version such as "snapshot" never reports an update.
func IsNewer(latest, current string) bool {
	latest, current = ensureV(latest), ensureV(current)
	if !semver.IsValid(latest) || !semver.IsValid(current) {
		return false
	}
	return semver.Compare(latest, current) > 0
}
