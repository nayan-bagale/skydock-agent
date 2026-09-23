package zmq

import (
	"context"
	"encoding/json"
	"os"

	"github.com/nayan-bagale/skydock-agent/internal/logger"
	"github.com/nayan-bagale/skydock-agent/internal/repository"
)

// Wire event names — keep in sync with apps/desktop/electron/agent-protocol.ts
const (
	EventAgentReady        = "agent:ready"
	EventAgentError        = "agent:error"
	EventAgentHeartbeat    = "agent:heartbeat"
	EventGetRecentActivity = "GET_RECENT_ACTIVITY"
	recentActivityLimit    = 50
)

// Deps is what inbound event handlers can use. Pass extra services here
// as you add events.
type Deps struct {
	Log     *logger.Logger
	Version string
	Roots   *repository.SyncRootRepository
	Files   *repository.FileRepository
}

// Register wires inbound event names to handlers.
//
// To add an event:
//  1. Add a name constant above (and in agent-protocol.ts)
//  2. Add s.On(...) below
//  3. Implement the handler next to the others
func Register(s *Server, d Deps) {
	s.On(EventAgentReady, handleAgentReady(d))
	s.On(EventAgentError, handleAgentError(d))
	s.On(EventGetRecentActivity, handleGetRecentActivity(d))
}

func handleAgentReady(d Deps) Handler {
	return func(ctx context.Context, peerID string, data json.RawMessage) (any, error) {
		roots, err := d.Roots.ListAll()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"ok":      true,
			"name":    "skydock-agent",
			"version": d.Version,
			"pid":     os.Getpid(),
			"roots":   len(roots),
		}, nil
	}
}

func handleAgentError(_ Deps) Handler {
	return func(ctx context.Context, peerID string, data json.RawMessage) (any, error) {
		return map[string]interface{}{
			"ok":    false,
			"error": string(data),
		}, nil
	}
}

func handleGetRecentActivity(d Deps) Handler {
	return func(ctx context.Context, peerID string, data json.RawMessage) (any, error) {
		records, err := d.Files.ListRecent(recentActivityLimit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"ok":      true,
			"records": records,
		}, nil
	}
}
