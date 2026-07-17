package storagetest

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bitwizeshift/go-cli/internal/storage"
)

// memFS is an in-memory [storage.FS]. Files are held in a flat map keyed by
// cleaned path; directories are tracked explicitly and marked implicitly as
// files are written beneath them.
type memFS struct {
	mu    sync.Mutex
	files map[string][]byte
	dirs  map[string]bool
}

func newMemFS() *memFS {
	return &memFS{
		files: map[string][]byte{},
		dirs:  map[string]bool{},
	}
}

// Open opens name for reading, returning [io/fs.ErrNotExist] when name is not a
// stored file.
func (m *memFS) Open(name string) (fs.File, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := filepath.Clean(name)
	data, ok := m.files[clean]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{
		info:   fileInfo{name: filepath.Base(clean), size: int64(len(data))},
		reader: bytes.NewReader(data),
	}, nil
}

// OpenWrite opens name for writing. The file is stored, and its parent
// directories marked, when the returned writer is closed.
func (m *memFS) OpenWrite(name string) (io.WriteCloser, error) {
	return &memWriter{fs: m, name: filepath.Clean(name)}, nil
}

// MkdirAll marks dir and all of its parents as directories.
func (m *memFS) MkdirAll(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.markDirs(filepath.Clean(dir))
	return nil
}

// Stat returns file information for name, returning [io/fs.ErrNotExist] when
// name names neither a stored file nor a known directory.
func (m *memFS) Stat(name string) (fs.FileInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := filepath.Clean(name)
	if data, ok := m.files[clean]; ok {
		return fileInfo{name: filepath.Base(clean), size: int64(len(data))}, nil
	}
	if m.isDir(clean) {
		return fileInfo{name: filepath.Base(clean), dir: true}, nil
	}
	return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
}

// ReadDir lists the immediate children of name, sorted by filename, returning
// [io/fs.ErrNotExist] when name is not a known directory.
func (m *memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := filepath.Clean(name)
	if !m.isDir(clean) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}

	seen := map[string]bool{}
	var entries []fs.DirEntry
	add := func(path string, dir bool) {
		if filepath.Dir(path) != clean || path == clean {
			return
		}
		base := filepath.Base(path)
		if seen[base] {
			return
		}
		seen[base] = true
		entries = append(entries, dirEntry{info: fileInfo{name: base, dir: dir}})
	}
	for path := range m.files {
		add(path, false)
	}
	for path := range m.dirs {
		add(path, true)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	return entries, nil
}

// Remove removes the file or empty directory named by name.
func (m *memFS) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := filepath.Clean(name)
	if _, ok := m.files[clean]; ok {
		delete(m.files, clean)
		return nil
	}
	if m.dirs[clean] {
		if m.hasChildren(clean) {
			return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrInvalid}
		}
		delete(m.dirs, clean)
		return nil
	}
	return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
}

// RemoveAll removes name and every file and directory beneath it. It is not an
// error for name to be absent.
func (m *memFS) RemoveAll(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	clean := filepath.Clean(name)
	prefix := clean + string(filepath.Separator)
	delete(m.files, clean)
	delete(m.dirs, clean)
	for path := range m.files {
		if strings.HasPrefix(path, prefix) {
			delete(m.files, path)
		}
	}
	for path := range m.dirs {
		if strings.HasPrefix(path, prefix) {
			delete(m.dirs, path)
		}
	}
	return nil
}

// markDirs records dir and each of its ancestors as directories.
func (m *memFS) markDirs(dir string) {
	for dir != "." && dir != string(filepath.Separator) && dir != "" {
		m.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

// isDir reports whether path is a known directory. The synthetic root "." is
// always a directory.
func (m *memFS) isDir(path string) bool {
	return path == "." || m.dirs[path]
}

// hasChildren reports whether any file or directory is stored beneath path.
func (m *memFS) hasChildren(path string) bool {
	prefix := path + string(filepath.Separator)
	for file := range m.files {
		if strings.HasPrefix(file, prefix) {
			return true
		}
	}
	for dir := range m.dirs {
		if strings.HasPrefix(dir, prefix) {
			return true
		}
	}
	return false
}

var _ storage.FS = (*memFS)(nil)

// memWriter buffers writes and, on close, stores the result into its [memFS] and
// marks the file's parent directories.
type memWriter struct {
	fs   *memFS
	name string
	buf  bytes.Buffer
}

func (mw *memWriter) Write(p []byte) (int, error) {
	return mw.buf.Write(p)
}

func (mw *memWriter) Close() error {
	mw.fs.mu.Lock()
	defer mw.fs.mu.Unlock()

	mw.fs.files[mw.name] = append([]byte(nil), mw.buf.Bytes()...)
	mw.fs.markDirs(filepath.Dir(mw.name))
	return nil
}

var _ io.WriteCloser = (*memWriter)(nil)

// memFile is a read-only handle to a stored file.
type memFile struct {
	info   fileInfo
	reader *bytes.Reader
}

func (mf *memFile) Stat() (fs.FileInfo, error) { return mf.info, nil }
func (mf *memFile) Read(p []byte) (int, error) { return mf.reader.Read(p) }
func (mf *memFile) Close() error               { return nil }

var _ fs.File = (*memFile)(nil)

// dirEntry adapts a [fileInfo] to [io/fs.DirEntry].
type dirEntry struct {
	info fileInfo
}

func (de dirEntry) Name() string               { return de.info.name }
func (de dirEntry) IsDir() bool                { return de.info.dir }
func (de dirEntry) Type() fs.FileMode          { return de.info.Mode().Type() }
func (de dirEntry) Info() (fs.FileInfo, error) { return de.info, nil }

var _ fs.DirEntry = dirEntry{}

// fileInfo is the [io/fs.FileInfo] for an in-memory file or directory.
type fileInfo struct {
	name string
	size int64
	dir  bool
}

func (fi fileInfo) Name() string { return fi.name }
func (fi fileInfo) Size() int64  { return fi.size }
func (fi fileInfo) Mode() fs.FileMode {
	if fi.dir {
		return fs.ModeDir | 0o755
	}
	return 0o644
}
func (fi fileInfo) ModTime() time.Time { return time.Time{} }
func (fi fileInfo) IsDir() bool        { return fi.dir }
func (fi fileInfo) Sys() any           { return nil }

var _ fs.FileInfo = fileInfo{}
