package zmq

import (
	"os"

	"github.com/nayan-bagale/skydock-agent/internal/config"
)

// Wire protocol v1 (event names live in events.go).

const ProtocolVersion = 1

const DefaultZMQAddr = config.DefaultZMQURL

func ResolveAddr() string {
	if v := config.Current().ZMQURL; v != "" {
		return v
	}
	if v := os.Getenv("SKYDOCK_AGENT_ZMQ_URL"); v != "" {
		return v
	}
	return DefaultZMQAddr
}
