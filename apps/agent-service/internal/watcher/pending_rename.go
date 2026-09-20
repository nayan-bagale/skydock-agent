package watcher

import (
	"time"

	internal "github.com/nayan-bagale/skydock-agent/internal"
)

// Unpaired Rename(old path) waits for a Create with the same inode. If none
// arrives before the settle deadline, the file is marked MISSING; the
// reconciler deletes the row after a successful scan.
type pendingRename struct {
	path     string
	inode    uint64
	device   uint64
	deadline time.Time
}

// clock returns the current time, or the injected clock used by tests.
func (w *Watcher) clock() time.Time {
	if w.now != nil {
		return w.now()
	}
	return time.Now()
}

// queueRename records that path disappeared via Rename and should become
// MISSING if the same inode is not observed again before the settle deadline.
func (w *Watcher) queueRename(path string, device, inode uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pendingRenames == nil {
		w.pendingRenames = make(map[string]pendingRename)
	}

	interval := w.settleInterval
	if interval <= 0 {
		interval = internal.RenameSettleInterval
	}

	w.pendingRenames[path] = pendingRename{
		path:     path,
		inode:    inode,
		device:   device,
		deadline: w.clock().Add(interval),
	}
}

// clearPendingForPath drops a pending rename keyed by the old path.
func (w *Watcher) clearPendingForPath(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.pendingRenames, path)
}

// clearPendingForInode drops pending renames whose identity matched a Create.
func (w *Watcher) clearPendingForInode(device, inode uint64) {
	if inode == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for path, pending := range w.pendingRenames {
		if pending.device == device && pending.inode == inode {
			delete(w.pendingRenames, path)
		}
	}
}

// expirePending marks leftover Rename paths MISSING when their inode never
// reappeared in the sync tree before the deadline. The reconciler deletes
// those rows after a later successful scan.
func (w *Watcher) expirePending(now time.Time) {
	if w.files == nil {
		return
	}

	w.mu.Lock()
	expired := make([]pendingRename, 0)
	for path, pending := range w.pendingRenames {
		if now.Before(pending.deadline) {
			continue
		}
		expired = append(expired, pending)
		delete(w.pendingRenames, path)
	}
	w.mu.Unlock()

	for _, pending := range expired {
		current, err := w.files.GetByInode(pending.device, pending.inode)
		if err != nil {
			w.log.Error("failed to resolve pending rename", "path", pending.path, "error", err)
			continue
		}
		if current != nil && current.Path != pending.path {
			continue
		}

		if err := w.files.MarkMissing(pending.path); err != nil {
			w.log.Error("failed to mark renamed file missing", "path", pending.path, "error", err)
		}
	}
}
