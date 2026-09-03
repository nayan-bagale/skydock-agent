package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nayan-bagale/skydock-agent/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(path string) (*gorm.DB, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if err := db.AutoMigrate(&models.FileRecord{}); err != nil {
		return fmt.Errorf("auto migrate database: %w", err)
	}

	return nil
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}
