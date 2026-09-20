package reconciler

import (
	"context"
	"time"
)

// DefaultInterval is the normal interval between filesystem reconciliation runs.
const DefaultInterval = 30 * time.Second

// Run starts the background reconciliation loop and returns when ctx is canceled.
// Reconcile is not run immediately here because the application performs the
// initial scan before starting the watcher and this service.
func (r *Reconciler) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.Reconcile(); err != nil {
				r.log.Error("filesystem reconciliation failed", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
