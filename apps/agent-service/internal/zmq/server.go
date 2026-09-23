package zmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-zeromq/zmq4"
	"github.com/nayan-bagale/skydock-agent/internal/logger"
)

type Server struct {
	addr     string
	log      *logger.Logger
	socket   zmq4.Socket
	peers    map[string]string
	peersMu  sync.RWMutex
	handlers map[string]Handler
	hMu      sync.RWMutex
	sendMu   sync.Mutex
}

func NewServer(addr string, log *logger.Logger) (*Server, error) {
	if addr == "" {
		addr = DefaultZMQAddr
	}
	return &Server{
		addr:     addr,
		log:      log,
		peers:    make(map[string]string),
		handlers: make(map[string]Handler),
	}, nil
}

func (s *Server) On(name string, handler Handler) {
	s.hMu.Lock()
	s.handlers[name] = handler
	s.hMu.Unlock()
}

func (s *Server) Close() error {
	if s.socket == nil {
		return nil
	}
	err := s.socket.Close()
	s.socket = nil
	return err
}

func (s *Server) Run(ctx context.Context) error {
	router := zmq4.NewRouter(ctx)
	s.socket = router

	if err := router.Listen(s.addr); err != nil {
		_ = router.Close()
		s.socket = nil
		return fmt.Errorf("bind ZMQ %s: %w", s.addr, err)
	}
	s.log.Info("ZMQ listening", "addr", s.addr)

	defer func() {
		_ = router.Close()
		s.socket = nil
	}()

	for {
		if ctx.Err() != nil {
			s.log.Info("ZMQ server stopping")
			return nil
		}

		msg, err := router.Recv()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			s.log.Error("ZMQ recv failed", "error", err)
			continue
		}
		if len(msg.Frames) < 2 {
			s.log.Warn("ZMQ message too short", "frames", len(msg.Frames))
			continue
		}

		identity := string(msg.Frames[0])
		payload := msg.Frames[len(msg.Frames)-1]
		peerID := peerKey(identity)
		s.registerPeer(peerID, identity)

		if err := s.handleInbound(ctx, peerID, payload); err != nil {
			s.log.Warn("handle inbound message", "error", err, "peer", peerID)
		}
	}
}

func peerKey(identity string) string {
	return identity
}

func (s *Server) registerPeer(peerID string, identity string) {
	s.peersMu.Lock()
	s.peers[peerID] = identity
	s.peersMu.Unlock()
}

func (s *Server) peerIdentity(peerID string) (string, bool) {
	s.peersMu.RLock()
	id, ok := s.peers[peerID]
	s.peersMu.RUnlock()
	return id, ok
}

func (s *Server) sendToPeer(identity string, payload []byte) error {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	msg := zmq4.NewMsgFrom([]byte(identity), []byte(""), payload)
	return s.socket.Send(msg)
}

func (s *Server) handleInbound(ctx context.Context, peerID string, payload []byte) error {
	env, err := UnmarshalEnvelope(payload)
	if err != nil {
		return err
	}
	if env.Kind != KindEvent {
		return fmt.Errorf("unexpected kind %q", env.Kind)
	}

	s.hMu.RLock()
	handler, ok := s.handlers[env.Name]
	s.hMu.RUnlock()
	if !ok {
		if env.ID != "" {
			ack, _ := MarshalAck(env.Name, map[string]interface{}{
				"ok":    false,
				"error": "unknown event",
			}, env.ID)
			identity, _ := s.peerIdentity(peerID)
			return s.sendToPeer(identity, ack)
		}
		return fmt.Errorf("no handler for %q", env.Name)
	}

	result, err := handler(ctx, peerID, env.Data)
	if env.ID == "" {
		return err
	}

	var ackData any
	if err != nil {
		ackData = map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		}
	} else {
		ackData = result
	}

	ack, err := MarshalAck(env.Name, ackData, env.ID)
	if err != nil {
		return err
	}
	identity, ok := s.peerIdentity(peerID)
	if !ok {
		return fmt.Errorf("peer %q disappeared", peerID)
	}
	return s.sendToPeer(identity, ack)
}

// ParseAckData is used when the desktop replies to agent-initiated requests (future use).
func ParseAckData(data json.RawMessage) (map[string]interface{}, error) {
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
