package repository

import (
	"fmt"
	"path/filepath"

	"github.com/nayan-bagale/skydock-agent/internal/models"
	"gorm.io/gorm"
)

type SyncRootRepository struct {
	db *gorm.DB
}

func NewSyncRootRepository(db *gorm.DB) *SyncRootRepository {
	return &SyncRootRepository{db: db}
}

func (r *SyncRootRepository) ListAll() ([]models.SyncRoot, error) {
	if err := r.requireDB(); err != nil {
		return nil, err
	}

	var roots []models.SyncRoot
	if err := r.db.Order("path ASC").Find(&roots).Error; err != nil {
		return nil, fmt.Errorf("list sync roots: %w", err)
	}
	return roots, nil
}

func (r *SyncRootRepository) ListEnabledPaths() ([]string, error) {
	if err := r.requireDB(); err != nil {
		return nil, err
	}

	var roots []models.SyncRoot
	if err := r.db.Where("enabled = ?", true).Order("path ASC").Find(&roots).Error; err != nil {
		return nil, fmt.Errorf("list enabled sync roots: %w", err)
	}

	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		paths = append(paths, root.Path)
	}

	return paths, nil
}

// Insert creates a new sync root. Duplicate paths are an error.
func (r *SyncRootRepository) Insert(root *models.SyncRoot) error {
	if err := r.requireDB(); err != nil {
		return err
	}

	if root == nil {
		return fmt.Errorf("sync root is nil")
	}

	if root.Path == "" {
		return fmt.Errorf("path is required")
	}

	if root.Name == "" {
		root.Name = filepath.Base(root.Path)
	}

	if err := r.db.Create(root).Error; err != nil {
		return fmt.Errorf("insert sync root: %w", err)
	}

	return nil
}

// Remove deletes the sync root at path. Missing rows are not an error.
func (r *SyncRootRepository) Remove(path string) error {
	if err := r.requireDB(); err != nil {
		return err
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	if err := r.db.Where("path = ?", path).Delete(&models.SyncRoot{}).Error; err != nil {
		return fmt.Errorf("remove sync root: %w", err)
	}

	return nil
}

// Update replaces fields on the sync root currently stored at path.
func (r *SyncRootRepository) Update(path string, root *models.SyncRoot) error {
	if err := r.requireDB(); err != nil {
		return err
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	if root == nil {
		return fmt.Errorf("sync root is nil")
	}

	if root.Path == "" {
		root.Path = path
	}

	if root.Name == "" {
		root.Name = filepath.Base(root.Path)
	}

	result := r.db.Model(&models.SyncRoot{}).Where("path = ?", path).Updates(map[string]interface{}{
		"path":      root.Path,
		"name":      root.Name,
		"enabled":   root.Enabled,
		"inode":     root.Inode,
		"device":    root.Device,
		"remote_id": root.RemoteID,
	})
	if result.Error != nil {
		return fmt.Errorf("update sync root: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update sync root: not found")
	}

	return nil
}

func (r *SyncRootRepository) requireDB() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("sync root repository is not initialized")
	}

	return nil
}
