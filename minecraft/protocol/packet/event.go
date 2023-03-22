package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// UsePlayerID ...
	// TODO: Figure out what UsePlayerID is for.
	UsePlayerID byte
	// Event is the event that is transmitted.
	Event protocol.Event
}

// ID ...
func (*Event) ID() uint32 {
	return IDEvent
}

// Marshal ...
func (pk *Event) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *Event) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *Event) marshal(r protocol.IO) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.EventType(&pk.Event)
	r.Uint8(&pk.UsePlayerID)
	pk.Event.Marshal(r)
}
