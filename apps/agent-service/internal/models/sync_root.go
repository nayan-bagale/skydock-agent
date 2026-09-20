package models

import "time"

type SyncRoot struct {
	ID uint `gorm:"primaryKey"`

	Path string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null"`

	Enabled bool `gorm:"not null;default:true"`

	Inode  uint64 `gorm:"index"`
	Device uint64 `gorm:"index"`

	RemoteID string `gorm:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
