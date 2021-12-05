package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SimpleEventCommandsEnabled = iota + 1
	SimpleEventCommandsDisabled
	SimpleEventUnlockWorldTemplateSettings
)

// SimpleEvent is sent by the server to send a 'simple event' to the client, meaning an event without any
// additional event data. The event is typically used by the client for telemetry.
type SimpleEvent struct {
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType int16
}

// ID ...
func (*SimpleEvent) ID() uint32 {
	return IDSimpleEvent
}

// Marshal ...
func (pk *SimpleEvent) Marshal(w *protocol.Writer) {
	w.Int16(&pk.EventType)
}

// Unmarshal ...
func (pk *SimpleEvent) Unmarshal(r *protocol.Reader) {
	r.Int16(&pk.EventType)
}
