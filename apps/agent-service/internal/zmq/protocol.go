package zmq

import "os"

// Wire protocol v1 (event names live in events.go).

const ProtocolVersion = 1

const DefaultZMQAddr = "tcp://127.0.0.1:17300"

func ResolveAddr() string {
	if v := os.Getenv("SKYDOCK_AGENT_ZMQ_URL"); v != "" {
		return v
	}
	return DefaultZMQAddr
}
