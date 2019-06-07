package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *SimpleEvent) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.EventType)
}

// Unmarshal ...
func (pk *SimpleEvent) Unmarshal(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.LittleEndian, &pk.EventType)
}
