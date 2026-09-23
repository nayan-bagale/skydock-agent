package zmq

import (
	"context"
	"encoding/json"
)

type Handler func(ctx context.Context, peerID string, data json.RawMessage) (any, error)
