package watcher

import (
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	constants "github.com/nayan-bagale/skydock-agent/internal"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
)

var lastEvent = make(map[string]time.Time)

type Event struct {
	Path string
	Op   fsnotify.Op
}

// HandleEvent dispatches a filesystem event after ignoring chmod, temps, and
// path-level debounce. Directory creates ingest the full subtree.
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
		if isNewDirectory(event.Name) {
			ingestDirectoryTree(w, event.Name)
		} else {
			HandleCreate(w, event)
		}
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
		log.Info("ignoring permission change event", "path", event.Name)
		return true
	}

	// Ignore temporary files (like editor backups)
	if constants.TemporaryFileSuffix != "" && strings.HasSuffix(event.Name, constants.TemporaryFileSuffix) {
		return true
	}

	if constants.IgnoredName(event.Name) {
		return true
	}

	// Debounce events that fire too quickly
	now := time.Now()
	if t, exists := lastEvent[event.Name]; exists {
		if now.Sub(t) < constants.DebounceInterval {
			return true
		}
	}
	lastEvent[event.Name] = now

	return false
}

func isNewDirectory(path string) bool {
	isDir, err := file.IsDirectory(path)
	return err == nil && isDir
}
