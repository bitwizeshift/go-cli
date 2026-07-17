package storage

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bitwizeshift/go-cli/internal/storage/userdir"
)

// AppStorage groups the writable roots an application uses, each scoped beneath
// an application-specific subdirectory of the corresponding host location.
type AppStorage struct {
	// Config holds user-editable configuration.
	Config *Storage

	// Cache holds regeneratable data that may be discarded at any time.
	Cache *Storage

	// Data holds durable, persistent application data.
	Data *Storage

	// Runtime holds ephemeral, session-scoped files that the operating system
	// may clear when the login session ends.
	Runtime *Storage
}

// NewAppStorage returns the [AppStorage] for appID, resolving each root from the
// host operating system's conventions and appending appID as a subdirectory.
//
// Resolution failure for a root is deferred rather than fatal: the affected root
// reports the failure from its own operations, so an application that never
// touches that root is unaffected.
func NewAppStorage(appID string) *AppStorage {
	return &AppStorage{
		Config:  rootStorage(os.UserConfigDir, appID),
		Cache:   rootStorage(os.UserCacheDir, appID),
		Data:    rootStorage(dataDir, appID),
		Runtime: rootStorage(runtimeDir, appID),
	}
}

// dataDir resolves the host's durable application-data directory.
func dataDir() (string, error) {
	return userdir.DataDir(runtime.GOOS, os.Getenv, os.UserHomeDir)
}

// runtimeDir resolves the host's ephemeral runtime directory.
func runtimeDir() (string, error) {
	return userdir.RuntimeDir(runtime.GOOS, os.Getenv, os.TempDir)
}

// rootStorage resolves a base directory through resolve and returns a [Storage]
// rooted at that directory joined with appID. A resolution error yields a
// [Storage] whose operations all report that error.
func rootStorage(resolve func() (string, error), appID string) *Storage {
	base, err := resolve()
	if err != nil {
		return New("", ErrFS(err))
	}
	return New(filepath.Join(base, appID), osFS{})
}

// ErrFS returns an [FS] whose every operation reports err. It backs a [Storage]
// whose base directory could not be resolved, deferring the failure to the point
// of use.
func ErrFS(err error) FS {
	return errFS{err: err}
}

// errFS is the [FS] returned by [ErrFS].
type errFS struct {
	err error
}

func (e errFS) Open(string) (fs.File, error)             { return nil, e.err }
func (e errFS) OpenWrite(string) (io.WriteCloser, error) { return nil, e.err }
func (e errFS) MkdirAll(string) error                    { return e.err }
func (e errFS) Stat(string) (fs.FileInfo, error)         { return nil, e.err }
func (e errFS) ReadDir(string) ([]fs.DirEntry, error)    { return nil, e.err }
func (e errFS) Remove(string) error                      { return e.err }
func (e errFS) RemoveAll(string) error                   { return e.err }

var _ FS = errFS{}
