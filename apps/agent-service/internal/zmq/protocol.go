package zmq

import "os"

// Wire protocol v1 (event names live in events.go).

const ProtocolVersion = 1

const DefaultZMQAddr = "ipc:///tmp/skydock-agent.sock"

func ResolveAddr() string {
	if v := os.Getenv("SKYDOCK_AGENT_ZMQ_URL"); v != "" {
		return v
	}
	return DefaultZMQAddr
}
