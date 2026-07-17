package storagetest

import (
	"github.com/bitwizeshift/go-cli/internal/storage"
)

// New returns an in-memory [storage.Storage] rooted at root. Reads and writes
// stay in memory, so storage-backed code can be exercised without touching the
// real filesystem.
func New(root string) *storage.Storage {
	return storage.New(root, newMemFS())
}

// NewAppStorage returns a [storage.AppStorage] whose four roots are backed by
// in-memory storage. The roots are isolated from one another and from the real
// filesystem, so writes to one are not visible through another.
func NewAppStorage() *storage.AppStorage {
	backend := newMemFS()
	return &storage.AppStorage{
		Config:  storage.New("config", backend),
		Cache:   storage.New("cache", backend),
		Data:    storage.New("data", backend),
		Runtime: storage.New("runtime", backend),
	}
}
