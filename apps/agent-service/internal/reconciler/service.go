package reconciler

import (
	"context"
	"time"
)

const DefaultInterval = 30 * time.Second

func (r *Reconciler) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.Reconcile(); err != nil && r.log != nil {
				r.log.Error("filesystem reconciliation failed", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
