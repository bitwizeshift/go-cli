package storage

import (
	"io"
	"io/fs"
	"os"
)

// osFS is the production [FS], delegating directly to the os package. It carries
// no state of its own; the paths it receives are already rooted by a [Storage].
type osFS struct{}

// Open opens name for reading.
func (osFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// OpenWrite opens name for writing, creating it if absent and truncating it
// otherwise.
func (osFS) OpenWrite(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
}

// MkdirAll creates dir and any missing parents.
func (osFS) MkdirAll(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// Stat returns file information for name.
func (osFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// ReadDir lists the directory named by name, sorted by filename.
func (osFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// Remove removes the file or empty directory named by name.
func (osFS) Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll removes name and any children it contains.
func (osFS) RemoveAll(name string) error {
	return os.RemoveAll(name)
}

var _ FS = osFS{}
