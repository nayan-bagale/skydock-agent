package watcher

import (
	"github.com/fsnotify/fsnotify"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
)

func HandleCreate(w *Watcher, event fsnotify.Event) {
	if w.log != nil {
		w.log.Info("file created", "path", event.Name)
	}

	if w.files == nil {
		return
	}

	meta, err := file.GetFileMetadata(event.Name)
	if err != nil {
		if w.log != nil {
			w.log.Error("failed to load created file metadata", "path", event.Name, "error", err)
		}
		return
	}

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
	if err != nil {
		if w.log != nil {
			w.log.Error("failed to load modified file metadata", "path", event.Name, "error", err)
		}
		return
	}

	if err := w.files.Upsert(meta); err != nil && w.log != nil {
		w.log.Error("failed to persist modified file", "path", event.Name, "error", err)
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
