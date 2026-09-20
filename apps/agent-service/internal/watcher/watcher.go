package watcher

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	internal "github.com/nayan-bagale/skydock-agent/internal"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
)

type Watcher struct {
	fsWatcher      *fsnotify.Watcher
	log            *logger.Logger
	files          *repository.FileRepository
	mu             sync.Mutex
	pendingRenames map[string]pendingRename
	settleInterval time.Duration
	now            func() time.Time
}

// New creates an fsnotify watcher and an empty pending-rename map used to pair
// Rename(old path) with a later Create(new path).
func New(log *logger.Logger, files *repository.FileRepository) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	return &Watcher{
		fsWatcher:      w,
		log:            log,
		files:          files,
		pendingRenames: make(map[string]pendingRename),
		settleInterval: internal.RenameSettleInterval,
		now:            time.Now,
	}, nil
}

// Watch registers path with the underlying filesystem watcher.
func (w *Watcher) Watch(path string) error {
	return w.fsWatcher.Add(path)
}

// Events returns the raw fsnotify event channel.
func (w *Watcher) Events() <-chan fsnotify.Event {
	return w.fsWatcher.Events
}

// Errors returns the raw fsnotify error channel.
func (w *Watcher) Errors() <-chan error {
	return w.fsWatcher.Errors
}

// Close stops the underlying filesystem watcher.
func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}

// StartWatcher reads filesystem events until the watcher is closed and expires
// unpaired rename entries on each event and on RenameSettleInterval.
func (w *Watcher) StartWatcher() {
	interval := w.settleInterval
	if interval <= 0 {
		interval = internal.RenameSettleInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-w.Events():
			if !ok {
				return
			}

			HandleEvent(w, event)
			w.expirePending(w.clock())

		case err, ok := <-w.Errors():
			if !ok {
				return
			}
			w.log.Error("watcher error", "error", err)

		case <-ticker.C:
			w.expirePending(w.clock())
		}
	}
}
