package watcher

import (
	"io/fs"
	"path/filepath"

	file "github.com/nayan-bagale/skydock-agent/internal/file"
)

// ingestDirectoryTree registers fsnotify watches on every subdirectory under root
// and persists metadata for all files and folders in the tree. Used when a
// directory appears in a sync root (e.g. moved in) so children are not missed.
func ingestDirectoryTree(w *Watcher, root string) {
	if w == nil || w.files == nil {
		return
	}

	isDir, err := file.IsDirectory(root)
	if err != nil {
		w.log.Error("failed to inspect path for directory ingest", "path", root, "error", err)
		return
	}
	if !isDir {
		return
	}

	w.log.Info("ingesting directory tree", "path", root)

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			w.log.Error("failed to access path during directory ingest", "path", path, "error", walkErr)
			return nil
		}

		if entry.IsDir() {
			if addErr := w.fsWatcher.Add(path); addErr != nil {
				w.log.Error("failed to add watcher during directory ingest", "path", path, "error", addErr)
			}
		} else if !file.IsValidFile(path) {
			return nil
		}

		meta, metaErr := file.GetFileMetadata(path)
		if metaErr != nil {
			w.log.Error("failed to load metadata during directory ingest", "path", path, "error", metaErr)
			return nil
		}

		persistObserved(w, meta)
		return nil
	})
	if err != nil {
		w.log.Error("directory ingest walk failed", "path", root, "error", err)
	}
}
