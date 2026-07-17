package storage

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
)

// FS is the filesystem backend a [Storage] reads and writes through. Reads go
// through [FS.Open]; every other method mutates the filesystem. The paths passed
// to an FS are already resolved against a [Storage] root, so an implementation
// treats them as ordinary host paths.
type FS interface {
	// Open opens name for reading.
	Open(name string) (fs.File, error)

	// OpenWrite opens name for writing, truncating any existing file and
	// creating it if absent.
	OpenWrite(name string) (io.WriteCloser, error)

	// MkdirAll creates dir and any missing parents.
	MkdirAll(dir string) error

	// Stat returns file information for name.
	Stat(name string) (fs.FileInfo, error)

	// ReadDir lists the directory named by name, sorted by filename.
	ReadDir(name string) ([]fs.DirEntry, error)

	// Remove removes the file or empty directory named by name.
	Remove(name string) error

	// RemoveAll removes name and any children it contains.
	RemoveAll(name string) error
}

// Storage is a writable filesystem root scoped beneath a single base directory.
// It satisfies [io/fs.FS] for read access and adds create, write, and remove
// operations for mutation. Names are slash-separated and must satisfy
// [io/fs.ValidPath]; a name that escapes the root is rejected with
// [io/fs.ErrInvalid].
//
// The base directory is created lazily: it comes into existence on the first
// [Storage.Create] or [Storage.WriteFile], along with any intermediate
// directories the written path requires.
type Storage struct {
	root    string
	backend FS
}

// New returns a [Storage] rooted at root that reads and writes through backend.
func New(root string, backend FS) *Storage {
	return &Storage{root: root, backend: backend}
}

// Open opens name for reading, satisfying [io/fs.FS]. It returns an
// [io/fs.PathError] wrapping [io/fs.ErrInvalid] when name is not a valid path.
func (s *Storage) Open(name string) (fs.File, error) {
	path, err := s.resolve("open", name)
	if err != nil {
		return nil, err
	}
	return s.backend.Open(path)
}

// ReadFile reads the entire contents of name.
func (s *Storage) ReadFile(name string) (read []byte, err error) {
	file, err := s.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	return io.ReadAll(file)
}

// Stat returns file information for name.
func (s *Storage) Stat(name string) (fs.FileInfo, error) {
	path, err := s.resolve("stat", name)
	if err != nil {
		return nil, err
	}
	return s.backend.Stat(path)
}

// ReadDir lists the directory named by name, sorted by filename.
func (s *Storage) ReadDir(name string) ([]fs.DirEntry, error) {
	path, err := s.resolve("readdir", name)
	if err != nil {
		return nil, err
	}
	return s.backend.ReadDir(path)
}

// Create opens name for writing, creating any parent directories that do not
// yet exist and truncating any file already present. The caller is responsible
// for closing the returned writer.
func (s *Storage) Create(name string) (io.WriteCloser, error) {
	path, err := s.resolve("create", name)
	if err != nil {
		return nil, err
	}
	if err := s.backend.MkdirAll(filepath.Dir(path)); err != nil {
		return nil, err
	}
	return s.backend.OpenWrite(path)
}

// WriteFile writes data to name, creating parent directories as needed and
// replacing any existing contents.
func (s *Storage) WriteFile(name string, data []byte) error {
	writer, err := s.Create(name)
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return errors.Join(err, writer.Close())
	}
	return writer.Close()
}

// Remove removes the file or empty directory named by name.
func (s *Storage) Remove(name string) error {
	path, err := s.resolve("remove", name)
	if err != nil {
		return err
	}
	return s.backend.Remove(path)
}

// RemoveAll removes name and any children it contains. It is not an error for
// name to be absent.
func (s *Storage) RemoveAll(name string) error {
	path, err := s.resolve("removeall", name)
	if err != nil {
		return err
	}
	return s.backend.RemoveAll(path)
}

// Sub returns a [Storage] rooted at dir beneath s, sharing the same backend.
func (s *Storage) Sub(dir string) (*Storage, error) {
	path, err := s.resolve("sub", dir)
	if err != nil {
		return nil, err
	}
	return New(path, s.backend), nil
}

// resolve validates name and joins it onto the root, returning an
// [io/fs.PathError] wrapping [io/fs.ErrInvalid] when name is not a valid,
// non-escaping path.
func (s *Storage) resolve(op, name string) (string, error) {
	if !fs.ValidPath(name) {
		return "", &fs.PathError{Op: op, Path: name, Err: fs.ErrInvalid}
	}
	return filepath.Join(s.root, filepath.FromSlash(name)), nil
}

var (
	_ fs.FS         = (*Storage)(nil)
	_ fs.ReadFileFS = (*Storage)(nil)
	_ fs.StatFS     = (*Storage)(nil)
	_ fs.ReadDirFS  = (*Storage)(nil)
)
