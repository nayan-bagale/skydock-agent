package reconciler

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	filepkg "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
)

type inodeKey struct {
	device uint64
	inode  uint64
}

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
	if log == nil {
		log = discardLogger{}
	}

	return &Reconciler{
		roots: append([]string(nil), roots...),
		files: files,
		log:   log,
	}
}

type discardLogger struct{}

func (discardLogger) Info(string, ...any)  {}
func (discardLogger) Error(string, ...any) {}

// Reconcile makes the database reflect the current filesystem state.
// A failed or incomplete scan stops the operation before vanished records are
// deleted, preventing temporary filesystem errors from causing false deletions.
func (r *Reconciler) Reconcile() error {
	if r == nil || r.files == nil {
		return fmt.Errorf("reconciler is not initialized")
	}

	observed, observedInodes, err := r.scan()
	if err != nil {
		return err
	}

	records, err := r.files.ListAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		if !isWithinRoots(record.Path, r.roots) {
			continue
		}
		if _, exists := observed[record.Path]; exists {
			continue
		}

		if record.Inode != 0 {
			if newPath, ok := observedInodes[inodeKey{device: record.Device, inode: record.Inode}]; ok {
				dest, err := r.files.GetByPath(newPath)
				if err != nil {
					return err
				}
				if dest != nil {
					meta := repository.MetaFromRecord(dest)
					if err := r.files.Relocate(record.Path, meta); err != nil {
						return err
					}
				}
				r.log.Info("reconciliation relocated missing path", "from", record.Path, "to", newPath)
				continue
			}
		}

		if err := r.files.Delete(record.Path); err != nil {
			return err
		}
		r.log.Info("reconciliation deleted missing file", "path", record.Path)
	}

	r.log.Info("filesystem reconciliation complete", "files", len(observed))
	return nil
}

// scan walks every sync root, then ApplyObserved for each file and directory.
// On walk failure it returns before callers delete vanished rows.
func (r *Reconciler) scan() (map[string]struct{}, map[inodeKey]string, error) {
	observed := make(map[string]struct{})
	observedInodes := make(map[inodeKey]string)
	now := time.Now()
	var metas []*filepkg.FileMeta

	for _, root := range r.roots {
		// WalkDir does not follow symbolic links, which prevents a symlink from
		// making reconciliation unexpectedly scan outside the configured root.
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && !filepkg.IsValidFile(path) {
				return nil
			}

			meta, err := filepkg.GetFileMetadata(path)
			if err != nil {
				return fmt.Errorf("read metadata for %s: %w", path, err)
			}
			meta.LastSeenAt = now
			metas = append(metas, meta)
			observed[path] = struct{}{}
			if meta.Inode != 0 {
				observedInodes[inodeKey{device: meta.Device, inode: meta.Inode}] = path
			}
			return nil
		})
		if err != nil {
			return nil, nil, fmt.Errorf("scan sync root %s: %w", root, err)
		}
	}

	for _, meta := range metas {
		if err := r.files.ApplyObserved(meta); err != nil {
			return nil, nil, err
		}
	}

	return observed, observedInodes, nil
}

// isWithinRoots reports whether path is a descendant of one of the sync roots.
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

// isParentPath reports whether rel from filepath.Rel escaped above the root.
func isParentPath(path string) bool {
	separator := string(filepath.Separator)
	return len(path) > 2 && path[:3] == ".."+separator
}
