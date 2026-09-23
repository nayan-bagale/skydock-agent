package zmq

import (
	"encoding/json"
	"fmt"
)

type Kind string

const (
	KindEvent Kind = "event"
	KindAck   Kind = "ack"
)

type Envelope struct {
	V    int             `json:"v"`
	Kind Kind            `json:"kind"`
	Name string          `json:"name"`
	Data json.RawMessage `json:"data,omitempty"`
	ID   string          `json:"id,omitempty"`
}

func (e *Envelope) Validate() error {
	if e.V != ProtocolVersion {
		return fmt.Errorf("unsupported protocol version %d", e.V)
	}
	if e.Kind != KindEvent && e.Kind != KindAck {
		return fmt.Errorf("invalid kind %q", e.Kind)
	}
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func MarshalEvent(name string, data any, id string) ([]byte, error) {
	return marshalEnvelope(KindEvent, name, data, id)
}

func MarshalAck(name string, data any, id string) ([]byte, error) {
	return marshalEnvelope(KindAck, name, data, id)
}

func marshalEnvelope(kind Kind, name string, data any, id string) ([]byte, error) {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	env := Envelope{
		V:    ProtocolVersion,
		Kind: kind,
		Name: name,
		Data: raw,
		ID:   id,
	}
	return json.Marshal(env)
}

func UnmarshalEnvelope(payload []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return Envelope{}, err
	}
	if err := env.Validate(); err != nil {
		return Envelope{}, err
	}
	return env, nil
}
