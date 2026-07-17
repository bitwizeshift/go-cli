package userdir

import (
	"errors"
	"path/filepath"
)

// ErrMissingLocalAppData indicates that %LOCALAPPDATA% was not defined when
// resolving a Windows storage directory.
var ErrMissingLocalAppData = errors.New("%LOCALAPPDATA% is not defined")

// DataDir returns the base directory for durable, user-specific application
// data on goos, resolving environment variables through getenv and the home
// directory through home:
//
//   - Unix: $XDG_DATA_HOME, or ~/.local/share when unset
//   - macOS: ~/Library/Application Support
//   - Windows: %LOCALAPPDATA%
//
// It returns [ErrMissingLocalAppData] on Windows when %LOCALAPPDATA% is unset,
// or the error from home when the home directory is required but unavailable.
func DataDir(goos string, getenv func(string) string, home func() (string, error)) (string, error) {
	switch goos {
	case "windows":
		dir := getenv("LOCALAPPDATA")
		if dir == "" {
			return "", ErrMissingLocalAppData
		}
		return filepath.Clean(dir), nil
	case "darwin", "ios":
		dir, err := home()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "Library", "Application Support"), nil
	default:
		if dir := getenv("XDG_DATA_HOME"); dir != "" {
			return filepath.Clean(dir), nil
		}
		dir, err := home()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, ".local", "share"), nil
	}
}

// RuntimeDir returns the base directory for ephemeral, session-scoped runtime
// files on goos, resolving environment variables through getenv and the per-user
// temporary directory through tempDir:
//
//   - Unix: $XDG_RUNTIME_DIR, falling back to tempDir when unset
//   - macOS and Windows: tempDir
//
// These files belong to the current login session and may be cleared by the
// operating system when it ends. The error result exists for symmetry with
// [DataDir] and is always nil.
func RuntimeDir(goos string, getenv func(string) string, tempDir func() string) (string, error) {
	if goos != "windows" && goos != "darwin" && goos != "ios" {
		if dir := getenv("XDG_RUNTIME_DIR"); dir != "" {
			return dir, nil
		}
	}
	return tempDir(), nil
}
