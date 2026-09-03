package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
	syncStatus "github.com/nayan-bagale/skydock-agent/internal/models"
)

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

	meta.SyncStatus = syncStatus.SyncStatusCreated

	if err := w.files.Upsert(meta); err != nil && w.log != nil {
		w.log.Error("failed to persist created file", "path", event.Name, "error", err)
	}
}

func HandleWrite(w *Watcher, event fsnotify.Event) {
	if w.log != nil {
		w.log.Info("modified file", "path", event.Name)
	}

	if w.files == nil {
		return
	}

	meta, err := file.GetFileMetadata(event.Name)
	meta.SyncStatus = syncStatus.SyncStatusModified
	if err != nil {
		w.log.Error("failed to load modified file metadata", "path", event.Name, "error", err)
		return
	}

	if err := w.files.Upsert(meta); err != nil && w.log != nil {
		w.log.Error("failed to persist modified file", "path", event.Name, "error", err)
	}
}

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

	meta := &file.FileMeta{
		Path:        record.Path,
		Name:        record.Name,
		Size:        record.Size,
		ModifiedAt:  record.ModifiedAt,
		IsDirectory: record.IsDirectory,
		Inode:       record.Inode,
		Device:      record.Device,
		Checksum:    record.Checksum,
		RemoteID:    record.RemoteID,
		SyncStatus:  syncStatus.SyncStatusMissing,
		LastSeenAt:  time.Now(),
	}

	if err := w.files.Upsert(meta); err != nil && w.log != nil {
		w.log.Error("failed to persist renamed file", "path", event.Name, "error", err)
	}
}

func HandleRemove(w *Watcher, event fsnotify.Event) {
	if w.log != nil {
		w.log.Info("file removed", "path", event.Name)
	}

	if w.files == nil {
		return
	}

	if err := w.files.Delete(event.Name); err != nil && w.log != nil {
		w.log.Error("failed to remove file record", "path", event.Name, "error", err)
	}
}
