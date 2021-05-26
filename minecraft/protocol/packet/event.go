package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EventAchievementAwarded = iota
	EventEntityInteract
	EventPortalBuilt
	EventPortalUsed
	EventMobKilled
	EventCauldronUsed
	EventPlayerDeath
	EventBossKilled
	EventAgentCommand
	EventAgentCreated
	EventBannerPatternRemoved
	EventCommandExecuted
	EventFishBucketed
	EventPlayerWaxedOrUnwaxedCopper = 25
)

// Event is sent by the server to send an event with additional data. It is typically sent to the client for
// telemetry reasons, much like the SimpleEvent packet.
type Event struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType int32
	// UsePlayerID ... TODO: Figure out what this is for.
	UsePlayerID byte
}

// ID ...
func (*Event) ID() uint32 {
	return IDEvent
}

// Marshal ...
func (pk *Event) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.EventType)
	w.Uint8(&pk.UsePlayerID)

	// TODO: Add fields for all Event types.
}

// Unmarshal ...
func (pk *Event) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.EventType)
	r.Uint8(&pk.UsePlayerID)

	// TODO: Add fields for all Event types.
}
