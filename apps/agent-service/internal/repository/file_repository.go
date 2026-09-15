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

// NewFileRepository returns a repository backed by the given GORM connection.
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Upsert inserts a record for meta.Path or replaces every column when that path
// already exists. Identity is not preserved across path changes; use Relocate.
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

// Delete removes the record for path. Missing rows are not an error.
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

// GetByPath returns the record stored at path, or nil when none exists.
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

// GetByInode returns the record for a device/inode pair. inode 0 is treated as
// unknown identity and never matches.
func (r *FileRepository) GetByInode(device, inode uint64) (*models.FileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("file repository is not initialized")
	}

	if inode == 0 {
		return nil, nil
	}

	var record models.FileRecord
	if err := r.db.Where("device = ? AND inode = ?", device, inode).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get file record by inode: %w", err)
	}

	return &record, nil
}

// Relocate updates the existing identity row to a new path in a transaction.
// RemoteID is preserved. A row already occupying the destination is removed
// only when it is the same inode or has no remote identity.
func (r *FileRepository) Relocate(oldPath string, meta *file.FileMeta) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if meta == nil {
		return fmt.Errorf("file metadata is nil")
	}

	if meta.Path == "" {
		return fmt.Errorf("path is required")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		identity, err := loadIdentity(tx, oldPath, meta.Device, meta.Inode)
		if err != nil {
			return err
		}
		if identity == nil {
			return fmt.Errorf("relocate file record: identity not found")
		}

		if meta.Path != identity.Path {
			var occupant models.FileRecord
			err := tx.Where("path = ?", meta.Path).First(&occupant).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return fmt.Errorf("relocate file record: %w", err)
			}
			if err == nil && occupant.ID != identity.ID {
				sameIdentity := occupant.Device == identity.Device && occupant.Inode == identity.Inode && occupant.Inode != 0
				if !sameIdentity && occupant.RemoteID != "" {
					return fmt.Errorf("cannot relocate: destination %s belongs to a different remote file", meta.Path)
				}
				if err := tx.Delete(&occupant).Error; err != nil {
					return fmt.Errorf("relocate file record: %w", err)
				}
			}
		}

		status := meta.SyncStatus
		if status == "" {
			status = models.SyncStatusMoved
		}

		updates := map[string]interface{}{
			"path":         meta.Path,
			"name":         meta.Name,
			"size":         meta.Size,
			"modified_at":  meta.ModifiedAt,
			"is_directory": meta.IsDirectory,
			"inode":        meta.Inode,
			"device":       meta.Device,
			"checksum":     meta.Checksum,
			"sync_status":  status,
			"last_seen_at": meta.LastSeenAt,
		}

		if err := tx.Model(identity).Updates(updates).Error; err != nil {
			return fmt.Errorf("relocate file record: %w", err)
		}

		return nil
	})
}

// loadIdentity finds the row to relocate, preferring oldPath and falling back
// to device+inode when the path was already updated by a Create-first event.
func loadIdentity(tx *gorm.DB, oldPath string, device, inode uint64) (*models.FileRecord, error) {
	var identity models.FileRecord

	if oldPath != "" {
		err := tx.Where("path = ?", oldPath).First(&identity).Error
		if err == nil {
			return &identity, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("relocate file record: %w", err)
		}
	}

	if inode == 0 {
		return nil, nil
	}

	err := tx.Where("device = ? AND inode = ?", device, inode).First(&identity).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("relocate file record: %w", err)
	}

	return &identity, nil
}

// ApplyObserved upserts or relocates a row so the database matches a file that
// currently exists on disk. New identities are CREATED. Path changes keep the
// same row and become MOVED (or MODIFIED when content also changed).
func (r *FileRepository) ApplyObserved(meta *file.FileMeta) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if meta == nil {
		return fmt.Errorf("file metadata is nil")
	}

	existing, err := r.GetByInode(meta.Device, meta.Inode)
	if err != nil {
		return err
	}

	if existing != nil && existing.Path != meta.Path {
		assignObservedStatus(meta, existing, true)
		return r.Relocate(existing.Path, meta)
	}

	byPath, err := r.GetByPath(meta.Path)
	if err != nil {
		return err
	}

	if byPath == nil {
		if meta.SyncStatus == "" {
			meta.SyncStatus = models.SyncStatusCreated
		}
		return r.Upsert(meta)
	}

	assignObservedStatus(meta, byPath, false)
	return r.Upsert(meta)
}

// assignObservedStatus copies RemoteID from existing and sets SyncStatus for a
// file seen on disk. Relocated rows become MOVED unless content changed.
func assignObservedStatus(meta *file.FileMeta, existing *models.FileRecord, relocated bool) {
	meta.RemoteID = existing.RemoteID
	contentChanged := existing.Checksum != meta.Checksum

	if contentChanged {
		meta.SyncStatus = models.SyncStatusModified
		return
	}

	if relocated {
		if existing.SyncStatus == models.SyncStatusMissing && existing.RemoteID == "" {
			meta.SyncStatus = models.SyncStatusCreated
			return
		}
		meta.SyncStatus = models.SyncStatusMoved
		return
	}

	if existing.SyncStatus == models.SyncStatusMissing {
		if existing.RemoteID != "" {
			meta.SyncStatus = models.SyncStatusSynced
			return
		}
		meta.SyncStatus = models.SyncStatusCreated
		return
	}

	meta.SyncStatus = existing.SyncStatus
}

// MetaFromRecord copies a persisted row into FileMeta for relocate/upsert calls.
func MetaFromRecord(record *models.FileRecord) *file.FileMeta {
	if record == nil {
		return nil
	}

	return &file.FileMeta{
		Path:        record.Path,
		Name:        record.Name,
		Size:        record.Size,
		ModifiedAt:  record.ModifiedAt,
		IsDirectory: record.IsDirectory,
		Inode:       record.Inode,
		Device:      record.Device,
		Checksum:    record.Checksum,
		RemoteID:    record.RemoteID,
		SyncStatus:  record.SyncStatus,
		LastSeenAt:  record.LastSeenAt,
	}
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

// RemoveMissing deletes the record for a path that was previously marked missing
// during reconciliation.
func (r *FileRepository) RemoveMissing(path string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("file repository is not initialized")
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	if err := r.db.Where("path = ?", path).Delete(&models.FileRecord{}).Error; err != nil {
		return fmt.Errorf("remove missing file record: %w", err)
	}

	return nil
}
