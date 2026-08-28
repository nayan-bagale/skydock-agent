package watcher

import (
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	internal "github.com/nayan-bagale/skydock-agent/internal"
	file "github.com/nayan-bagale/skydock-agent/internal/file"
)

type Event struct {
	Path string
	Op   fsnotify.Op
}

func HandleEvent(w *Watcher, event fsnotify.Event) *Event {

	isIgnored := isIgnoredEvent(event)

	if isIgnored {
		return nil
	}

	// Handle file/directory creation
	if event.Has(fsnotify.Create) {
		isDir, err := file.IsDirectory(event.Name)
		if err != nil {
			if w.log != nil {
				w.log.Error("failed to inspect path", "path", event.Name, "error", err)
			}
			return nil
		}

		if isDir {
			if w.log != nil {
				w.log.Info("add watcher dynamically to new directory", "path", event.Name)
				err := w.fsWatcher.Add(event.Name)
				if err != nil {
					if w.log != nil {
						w.log.Error("failed to add watcher to new directory", "path", event.Name, "error", err)
					}
				}
			}
		} else {
			if w.log != nil {
				w.log.Info("run configured command for file", "path", event.Name)
			}
		}
	}

	// Handle file modifications
	if event.Has(fsnotify.Write) {
		if w.log != nil {
			w.log.Info("modified file", "path", event.Name)
		}
	}

	return &Event{
		Path: event.Name,
		Op:   event.Op,
	}
}

func isIgnoredEvent(event fsnotify.Event) bool {
	// Ignore permission change events
	if event.Has(fsnotify.Chmod) {
		return true
	}

	// Ignore temporary files (like editor backups)
	if strings.HasSuffix(event.Name, internal.TemporaryFileSuffix) {
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
