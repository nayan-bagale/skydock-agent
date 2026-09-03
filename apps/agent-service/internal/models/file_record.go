package models

import "time"

type FileRecord struct {
	ID uint `gorm:"primaryKey"`

	Path string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null"`

	Size        int64     `gorm:"not null"`
	ModifiedAt  time.Time `gorm:"not null"`
	IsDirectory bool      `gorm:"not null"`

	Inode  uint64 `gorm:"index"`
	Device uint64 `gorm:"index"`

	Checksum string `gorm:"index"`
	RemoteID string `gorm:"index"`

	// SYNCED, CREATED, MODIFIED, MOVED, DELETED, etc.
	SyncStatus string `gorm:"index"`
	// When the file was last observed during reconciliation.
	LastSeenAt time.Time `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
