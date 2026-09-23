package zmq

import (
	"fmt"
)

func (s *Server) Emit(peerID string, name string, data any) error {
	if s == nil || s.socket == nil {
		return fmt.Errorf("zmq server is not running")
	}
	payload, err := MarshalEvent(name, data, "")
	if err != nil {
		return err
	}
	identity, ok := s.peerIdentity(peerID)
	if !ok {
		return fmt.Errorf("unknown peer %q", peerID)
	}
	return s.sendToPeer(identity, payload)
}

func (s *Server) EmitAll(name string, data any) error {
	if s == nil || s.socket == nil {
		return fmt.Errorf("zmq server is not running")
	}
	payload, err := MarshalEvent(name, data, "")
	if err != nil {
		return err
	}

	s.peersMu.RLock()
	peers := make([]string, 0, len(s.peers))
	for _, id := range s.peers {
		peers = append(peers, id)
	}
	s.peersMu.RUnlock()

	var lastErr error
	for _, identity := range peers {
		if err := s.sendToPeer(identity, payload); err != nil {
			lastErr = err
			s.log.Warn("emit to peer failed", "error", err)
		}
	}
	return lastErr
}
