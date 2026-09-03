package repository

import (
	"fmt"

	"github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Upsert(meta *file.FileMeta) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if meta == nil {
		return fmt.Errorf("file metadata is nil")
	}

	record := models.FileRecord{
		Path:        meta.Path,
		Name:        meta.Name,
		Size:        meta.Size,
		ModifiedAt:  meta.ModifiedAt,
		IsDirectory: meta.IsDirectory,
		Inode:       meta.Inode,
		Device:      meta.Device,
		Checksum:    meta.Checksum,
		RemoteID:    meta.RemoteID,
		SyncStatus:  meta.SyncStatus,
		LastSeenAt:  meta.LastSeenAt,
	}

	if err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}},
		UpdateAll: true,
	}).Create(&record).Error; err != nil {
		return fmt.Errorf("upsert file record: %w", err)
	}

	return nil
}

func (r *FileRepository) Delete(path string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	if err := r.db.Where("path = ?", path).Delete(&models.FileRecord{}).Error; err != nil {
		return fmt.Errorf("delete file record: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByPath(path string) (*models.FileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("file repository is not initialized")
	}

	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	var record models.FileRecord
	if err := r.db.Where("path = ?", path).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get file record by path: %w", err)
	}

	return &record, nil
}

// ListAll returns all records so the reconciler can identify paths that no
// longer exist in the filesystem.
func (r *FileRepository) ListAll() ([]models.FileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("file repository is not initialized")
	}

	var records []models.FileRecord
	if err := r.db.Order("path ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list file records: %w", err)
	}

	return records, nil
}

// MarkMissing keeps the record for sync/history purposes while recording that
// the corresponding local path was not observed during reconciliation.
func (r *FileRepository) MarkMissing(path string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	updates := map[string]interface{}{
		"sync_status": models.SyncStatusMissing,
	}
	if err := r.db.Model(&models.FileRecord{}).Where("path = ?", path).Updates(updates).Error; err != nil {
		return fmt.Errorf("mark file record missing: %w", err)
	}

	return nil
}
