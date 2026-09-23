package zmq

import (
	"context"
	"time"
)

const HeartbeatInterval = 5 * time.Second

// StartPublishers launches background emits. Non-blocking.
//
// To add an outbound event:
//  1. Add a name constant in events.go (and agent-protocol.ts)
//  2. Add go publishX(ctx, s, d) below
//  3. Implement publishX next to the others
func StartPublishers(ctx context.Context, s *Server, d Deps) {
	go publishHeartbeat(ctx, s, d)
}

func publishHeartbeat(ctx context.Context, s *Server, d Deps) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if err := s.EmitAll(EventAgentHeartbeat, map[string]interface{}{
				"ts": t.UTC().Format(time.RFC3339),
			}); err != nil {
				d.Log.Debug("heartbeat emit", "error", err)
			}
		}
	}
}
