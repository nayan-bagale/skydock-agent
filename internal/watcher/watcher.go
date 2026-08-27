package watcher

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	internal "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
)

var lastEvent = make(map[string]time.Time)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	log       *logger.Logger
}

func New(log *logger.Logger) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	return &Watcher{
		fsWatcher: w,
		log:       log,
	}, nil
}

func (w *Watcher) Watch(path string) error {
	return w.fsWatcher.Add(path)
}

func (w *Watcher) Events() <-chan fsnotify.Event {
	return w.fsWatcher.Events
}

func (w *Watcher) Errors() <-chan error {
	return w.fsWatcher.Errors
}

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}

func (w *Watcher) StartWatcher() {
	for {
		select {

		// File system events
		case event, ok := <-w.Events():
			if !ok {
				return
			}

			// Ignore permission change events
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// Ignore temporary files (like editor backups)
			if strings.HasSuffix(event.Name, internal.TemporaryFileSuffix) {
				continue
			}

			// Debounce events that fire too quickly
			now := time.Now()
			if t, exists := lastEvent[event.Name]; exists {
				if now.Sub(t) < internal.DebounceInterval {
					continue
				}
			}
			lastEvent[event.Name] = now

			// Handle file/directory creation
			if event.Has(fsnotify.Create) {
				isDir, err := IsDirectory(event.Name)
				if err != nil {
					if w.log != nil {
						w.log.Error("failed to inspect path", "path", event.Name, "error", err)
					}
					continue
				}

				if isDir {
					if w.log != nil {
						w.log.Info("add watcher dynamically to new directory", "path", event.Name)
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

		// Error handling
		case err, ok := <-w.Errors():
			if !ok {
				return
			}
			if w.log != nil {
				w.log.Error("watcher error", "error", err)
			}
		}
	}
}

func (w *Watcher) Initialize() {
	watch, err := New(w.log)
	if err != nil {
		if w.log != nil {
			w.log.Error("failed to create watcher", "error", err)
		}
		return
	}
	defer watch.Close()

	dirs, err := GetAllDirs(internal.Directories)
	if err != nil {
		if w.log != nil {
			w.log.Error("failed to get directories", "error", err)
		}
		return
	}

	for _, dir := range dirs {
		err := watch.Watch(dir)
		if err != nil {
			if w.log != nil {
				w.log.Error("failed to watch directory", "path", dir, "error", err)
			}
			return
		}
	}

	go watch.StartWatcher()
}
