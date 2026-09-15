package watcher

import (
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	internal "github.com/nayan-bagale/skydock-agent/internal"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
)

var lastEvent = make(map[string]time.Time)

type Event struct {
	Path string
	Op   fsnotify.Op
}

// HandleEvent dispatches a filesystem event after ignoring chmod, temps, and
// path-level debounce. Directory creates only attach a new watch.
func HandleEvent(w *Watcher, event fsnotify.Event) {

	isIgnored := isIgnoredEvent(event, w.log)

	// w.log.Info(
	// 	"Handle Event",
	// 	"Path", event.Name,
	// 	event.Op.String(),
	// )

	if isIgnored {
		return
	}

	// Handle file/directory creation
	if event.Has(fsnotify.Create) {
		isAttached := attachWatcherToNewDirectory(w, event)

		if isAttached {
			return
		}

		HandleCreate(w, event)

	}

	// Handle file modifications
	if event.Has(fsnotify.Write) {
		HandleWrite(w, event)
	}

	if event.Has(fsnotify.Rename) {
		HandleRename(w, event)
	}

	if event.Has(fsnotify.Remove) {
		HandleRemove(w, event)
	}

}

// isIgnoredEvent returns true for chmod, editor temp files, .DS_Store, and
// duplicate events on the same path within DebounceInterval.
func isIgnoredEvent(event fsnotify.Event, log *logger.Logger) bool {
	// Ignore permission change events
	if event.Has(fsnotify.Chmod) {
		if log != nil {
			log.Info("ignoring permission change event", "path", event.Name)
		}
		return true
	}

	// Ignore temporary files (like editor backups)
	if strings.HasSuffix(event.Name, internal.TemporaryFileSuffix) {
		return true
	}

	// Ignore .DS_Store files
	if strings.HasSuffix(event.Name, internal.DS_StoreFileName) {
		return true
	}

	// Debounce events that fire too quickly
	now := time.Now()
	if t, exists := lastEvent[event.Name]; exists {
		if now.Sub(t) < internal.DebounceInterval {
			return true
		}
	}
	lastEvent[event.Name] = now

	return false
}

// attachWatcherToNewDirectory starts watching a newly created directory and
// returns true when the event was a directory (so it is not stored as a file).
func attachWatcherToNewDirectory(w *Watcher, event fsnotify.Event) bool {
	isDir, err := file.IsDirectory(event.Name)
	if err != nil {
		if w.log != nil {
			w.log.Error("failed to inspect path", "path", event.Name, "error", err)
		}
		return false
	}

	if !isDir {
		return false
	}

	if w.log != nil {
		w.log.Info("add watcher dynamically to new directory", "path", event.Name)
	}
	wErr := w.fsWatcher.Add(event.Name)
	if wErr != nil {
		if w.log != nil {
			w.log.Error("failed to add watcher to new directory", "path", event.Name, "error", wErr)
		}
		return false
	}
	return true
}
