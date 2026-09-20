package watcher

import (
	"github.com/fsnotify/fsnotify"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
)

// HandleCreate records a path that appeared. Same device+inode at a new path
// is treated as a move, not a new file.
func HandleCreate(w *Watcher, event fsnotify.Event) {
	w.log.Info("file created", "path", event.Name)

	if w.files == nil {
		return
	}

	meta, err := file.GetFileMetadata(event.Name)
	if err != nil {
		w.log.Error("failed to load created file metadata", "path", event.Name, "error", err)
		return
	}

	persistObserved(w, meta)
}

// HandleWrite refreshes metadata for an existing path and sets MODIFIED when
// the checksum changed.
func HandleWrite(w *Watcher, event fsnotify.Event) {
	w.log.Info("modified file", "path", event.Name)

	if w.files == nil {
		return
	}

	meta, err := file.GetFileMetadata(event.Name)
	if err != nil {
		w.log.Error("failed to load modified file metadata", "path", event.Name, "error", err)
		return
	}

	persistObserved(w, meta)
}

// HandleRename treats event.Name as the old path. It does not mark MISSING
// immediately; Create may still report the new path inside the settle window.
func HandleRename(w *Watcher, event fsnotify.Event) {
	w.log.Info("renamed file", "path", event.Name)

	if w.files == nil {
		return
	}

	record, err := w.files.GetByPath(event.Name)
	if err != nil {
		w.log.Error("failed to load renamed file metadata", "path", event.Name, "error", err)
		return
	}
	if record == nil {
		w.clearPendingForPath(event.Name)
		return
	}

	w.queueRename(record.Path, record.Device, record.Inode)
}

// HandleRemove deletes the local row for a true delete, distinct from rename.
func HandleRemove(w *Watcher, event fsnotify.Event) {
	w.log.Info("file removed", "path", event.Name)

	if w.files == nil {
		return
	}

	w.clearPendingForPath(event.Name)

	if err := w.files.Delete(event.Name); err != nil {
		w.log.Error("failed to remove file record", "path", event.Name, "error", err)
	}
}

// persistObserved writes observed file metadata and cancels any pending rename
// for that inode.
func persistObserved(w *Watcher, meta *file.FileMeta) {
	if meta == nil {
		return
	}

	if err := w.files.ApplyObserved(meta); err != nil {
		w.log.Error("failed to persist observed file", "path", meta.Path, "error", err)
		return
	}

	w.clearPendingForPath(meta.Path)
	w.clearPendingForInode(meta.Device, meta.Inode)
}
