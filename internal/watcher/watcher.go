package watcher

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
)

var lastEvent = make(map[string]time.Time)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	log       *logger.Logger
	files     *repository.FileRepository
}

func New(log *logger.Logger, files *repository.FileRepository) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	return &Watcher{
		fsWatcher: w,
		log:       log,
		files:     files,
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

			HandleEvent(w, event)

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
