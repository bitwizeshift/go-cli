package userdir_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/storage/userdir"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDataDir(t *testing.T) {
	t.Parallel()

	homeErr := errors.New("home unavailable")

	testCases := []struct {
		name    string
		goos    string
		env     map[string]string
		home    string
		homeErr error
		want    string
		wantErr error
	}{
		{
			name: "UnixWithXDGDataHome",
			goos: "linux",
			env:  map[string]string{"XDG_DATA_HOME": "/xdg/data"},
			home: "/home/user",
			want: filepath.Clean("/xdg/data"),
		},
		{
			name: "UnixWithoutXDGDataHome",
			goos: "linux",
			env:  map[string]string{},
			home: "/home/user",
			want: filepath.Clean("/home/user/.local/share"),
		},
		{
			name:    "UnixHomeError",
			goos:    "linux",
			env:     map[string]string{},
			homeErr: homeErr,
			wantErr: homeErr,
		},
		{
			name: "Darwin",
			goos: "darwin",
			env:  map[string]string{},
			home: "/Users/user",
			want: filepath.Clean("/Users/user/Library/Application Support"),
		},
		{
			name:    "DarwinHomeError",
			goos:    "darwin",
			env:     map[string]string{},
			homeErr: homeErr,
			wantErr: homeErr,
		},
		{
			name: "Windows",
			goos: "windows",
			env:  map[string]string{"LOCALAPPDATA": `C:\Users\user\AppData\Local`},
			want: filepath.Clean(`C:\Users\user\AppData\Local`),
		},
		{
			name:    "WindowsMissingLocalAppData",
			goos:    "windows",
			env:     map[string]string{},
			wantErr: userdir.ErrMissingLocalAppData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			getenv := func(key string) string { return tc.env[key] }
			home := func() (string, error) { return tc.home, tc.homeErr }

			// Act
			dir, err := userdir.DataDir(tc.goos, getenv, home)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("userdir.DataDir(...) error = %v, want %v", got, want)
			}
			if got, want := dir, tc.want; got != want {
				t.Errorf("userdir.DataDir(...) = %q, want %q", got, want)
			}
		})
	}
}

func TestRuntimeDir(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		goos    string
		env     map[string]string
		tempDir string
		want    string
	}{
		{
			name:    "UnixWithXDGRuntimeDir",
			goos:    "linux",
			env:     map[string]string{"XDG_RUNTIME_DIR": "/run/user/1000"},
			tempDir: "/tmp",
			want:    "/run/user/1000",
		},
		{
			name:    "UnixWithoutXDGRuntimeDir",
			goos:    "linux",
			env:     map[string]string{},
			tempDir: "/tmp",
			want:    "/tmp",
		},
		{
			name:    "DarwinIgnoresXDGRuntimeDir",
			goos:    "darwin",
			env:     map[string]string{"XDG_RUNTIME_DIR": "/run/user/1000"},
			tempDir: "/var/folders/tmp",
			want:    "/var/folders/tmp",
		},
		{
			name:    "Windows",
			goos:    "windows",
			env:     map[string]string{},
			tempDir: `C:\Temp`,
			want:    `C:\Temp`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			getenv := func(key string) string { return tc.env[key] }
			tempDir := func() string { return tc.tempDir }

			// Act
			dir, err := userdir.RuntimeDir(tc.goos, getenv, tempDir)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("userdir.RuntimeDir(...) error = %v, want %v", got, want)
			}
			if got, want := dir, tc.want; got != want {
				t.Errorf("userdir.RuntimeDir(...) = %q, want %q", got, want)
			}
		})
	}
}
