package reconciler

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	filepkg "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/models"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
)

type Reconciler struct {
	roots []string
	files *repository.FileRepository
	log   logger
}

type logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

// New creates a reconciler for the configured sync roots.
// The logger is intentionally an interface so the reconciliation logic stays
// independent from the concrete logging implementation and remains easy to test.
func New(roots []string, files *repository.FileRepository, log logger) *Reconciler {
	return &Reconciler{
		roots: append([]string(nil), roots...),
		files: files,
		log:   log,
	}
}

// Reconcile makes the database reflect the current filesystem state.
// A failed or incomplete scan stops the operation before missing records are
// marked, preventing temporary filesystem errors from causing false deletions.
func (r *Reconciler) Reconcile() error {
	if r == nil || r.files == nil {
		return fmt.Errorf("reconciler is not initialized")
	}

	observed, err := r.scan()
	if err != nil {
		return err
	}

	records, err := r.files.ListAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		// Only reconcile records owned by this service. This protects records
		// that may belong to another sync root or future feature.
		if !isWithinRoots(record.Path, r.roots) {
			continue
		}
		if _, exists := observed[record.Path]; exists {
			continue
		}

		if err := r.files.MarkMissing(record.Path); err != nil {
			return err
		}
		if r.log != nil {
			r.log.Info("reconciliation marked file missing", "path", record.Path)
		}
	}

	if r.log != nil {
		r.log.Info("filesystem reconciliation complete", "files", len(observed))
	}
	return nil
}

func (r *Reconciler) scan() (map[string]struct{}, error) {
	observed := make(map[string]struct{})
	now := time.Now()

	for _, root := range r.roots {
		// WalkDir does not follow symbolic links, which prevents a symlink from
		// making reconciliation unexpectedly scan outside the configured root.
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !filepkg.IsValidFile(path) {
				return nil
			}

			meta, err := filepkg.GetFileMetadata(path)
			if err != nil {
				return fmt.Errorf("read metadata for %s: %w", path, err)
			}
			meta.LastSeenAt = now

			record, err := r.files.GetByPath(path)
			if err != nil {
				return err
			}
			if record == nil {
				meta.SyncStatus = models.SyncStatusSynced
			} else {
				// Reconciliation refreshes local metadata without overwriting
				// remote identity or a pending sync state.
				meta.RemoteID = record.RemoteID
				meta.SyncStatus = record.SyncStatus
			}

			if err := r.files.Upsert(meta); err != nil {
				return err
			}
			observed[path] = struct{}{}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scan sync root %s: %w", root, err)
		}
	}

	return observed, nil
}

func isWithinRoots(path string, roots []string) bool {
	path = filepath.Clean(path)
	for _, root := range roots {
		root = filepath.Clean(root)
		rel, err := filepath.Rel(root, path)
		if err == nil && rel != ".." && rel != "." && !isParentPath(rel) {
			return true
		}
	}
	return false
}

func isParentPath(path string) bool {
	separator := string(filepath.Separator)
	return len(path) > 2 && path[:3] == ".."+separator
}
