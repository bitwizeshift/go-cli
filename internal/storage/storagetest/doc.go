// Package storagetest provides in-memory [storage.Storage] doubles for
// exercising storage-backed code without touching the real filesystem.
//
// [New] yields a single in-memory root and [NewAppStorage] yields a full
// application bundle whose four roots are isolated from one another. Reads and
// writes stay in process memory, so a consumer writes through a root and reads
// the same data back to make assertions.
package storagetest
