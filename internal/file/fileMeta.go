package file

import "time"

type FileMeta struct {
	Path        string
	Name        string
	Size        int64
	ModifiedAt  time.Time
	IsDirectory bool

	// Identity
	Inode  uint64 // macOS/Linux
	Device uint64

	// Content
	Checksum string

	// Sync
	RemoteID   string
	SyncStatus string
}
