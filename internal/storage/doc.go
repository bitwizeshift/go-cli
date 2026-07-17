// Package storage provides writable, per-application filesystem roots modeled
// on the sandboxed storage of mobile applications.
//
// An [AppStorage] groups the four roots an application typically needs — user
// configuration, regeneratable cache, durable data, and ephemeral runtime
// files — each resolved from the host operating system's conventions and scoped
// beneath an application-specific subdirectory. Every root is a [Storage]: a
// writable handle that satisfies [io/fs.FS] for reads and adds create, write,
// and remove operations. Directories are created lazily, on first write.
//
// [Storage] reads and writes through an [FS] backend. Production uses one backed
// by the os package; tests substitute an in-memory fake, so callers holding a
// [Storage] can be exercised without touching the real filesystem.
package storage
