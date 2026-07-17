package cli

import (
	"context"
	"io"
	"io/fs"

	"github.com/bitwizeshift/go-cli/internal/clictx"
	"github.com/bitwizeshift/go-cli/internal/storage"
)

// Storage is a writable application-storage root. It satisfies [io/fs.FS] for
// reading and adds create, write, and remove operations for mutation. Names are
// slash-separated and must satisfy [io/fs.ValidPath]; a name that escapes the
// root is rejected with [io/fs.ErrInvalid].
//
// The backing directory is created lazily, on the first [Storage.Create] or
// [Storage.WriteFile].
type Storage struct {
	impl *storage.Storage
}

// newStorage wraps an internal storage root.
func newStorage(impl *storage.Storage) *Storage {
	return &Storage{impl: impl}
}

// Open opens name for reading, satisfying [io/fs.FS].
func (s *Storage) Open(name string) (fs.File, error) {
	return s.impl.Open(name)
}

// ReadFile reads the entire contents of name.
func (s *Storage) ReadFile(name string) ([]byte, error) {
	return s.impl.ReadFile(name)
}

// Stat returns file information for name.
func (s *Storage) Stat(name string) (fs.FileInfo, error) {
	return s.impl.Stat(name)
}

// ReadDir lists the directory named by name, sorted by filename.
func (s *Storage) ReadDir(name string) ([]fs.DirEntry, error) {
	return s.impl.ReadDir(name)
}

// Create opens name for writing, creating any parent directories that do not
// yet exist and truncating any file already present. The caller closes the
// returned writer.
func (s *Storage) Create(name string) (io.WriteCloser, error) {
	return s.impl.Create(name)
}

// WriteFile writes data to name, creating parent directories as needed and
// replacing any existing contents.
func (s *Storage) WriteFile(name string, data []byte) error {
	return s.impl.WriteFile(name, data)
}

// Remove removes the file or empty directory named by name.
func (s *Storage) Remove(name string) error {
	return s.impl.Remove(name)
}

// RemoveAll removes name and any children it contains. It is not an error for
// name to be absent.
func (s *Storage) RemoveAll(name string) error {
	return s.impl.RemoveAll(name)
}

// Sub returns a [Storage] rooted at dir beneath s.
func (s *Storage) Sub(dir string) (*Storage, error) {
	sub, err := s.impl.Sub(dir)
	if err != nil {
		return nil, err
	}
	return newStorage(sub), nil
}

// AppStorage groups the writable storage roots available to a command, each
// scoped beneath an application-specific directory resolved from the host
// operating system's conventions.
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

// StorageFrom returns the [AppStorage] carried by the application's context, or
// nil when the context carries none (for example, outside a running command).
func StorageFrom(ctx context.Context) *AppStorage {
	app := clictx.Storage(ctx)
	if app == nil {
		return nil
	}
	return &AppStorage{
		Config:  newStorage(app.Config),
		Cache:   newStorage(app.Cache),
		Data:    newStorage(app.Data),
		Runtime: newStorage(app.Runtime),
	}
}

var (
	_ fs.FS         = (*Storage)(nil)
	_ fs.ReadFileFS = (*Storage)(nil)
	_ fs.StatFS     = (*Storage)(nil)
	_ fs.ReadDirFS  = (*Storage)(nil)
)
