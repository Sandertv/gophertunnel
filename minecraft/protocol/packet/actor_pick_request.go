package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ActorPickRequest is sent by the client when it tries to pick an entity, so that it gets a spawn egg which
// can spawn that entity.
type ActorPickRequest struct {
	// EntityUniqueID is the unique ID of the entity that was attempted to be picked. The server must find the
	// type of that entity and provide the correct spawn egg to the player.
	EntityUniqueID int64
	// HotBarSlot is the held hot bar slot of the player at the time of trying to pick the entity. If empty,
	// the resulting spawn egg should be put into this slot.
	HotBarSlot byte
	// WithData is true if the pick request requests the entity metadata.
	WithData bool
}

// ID ...
func (*ActorPickRequest) ID() uint32 {
	return IDActorPickRequest
}

// Marshal ...
func (pk *ActorPickRequest) Marshal(w *protocol.Writer) {
	w.Int64(&pk.EntityUniqueID)
	w.Uint8(&pk.HotBarSlot)
	w.Bool(&pk.WithData)
}

// Unmarshal ...
func (pk *ActorPickRequest) Unmarshal(r *protocol.Reader) {
	r.Int64(&pk.EntityUniqueID)
	r.Uint8(&pk.HotBarSlot)
	r.Bool(&pk.WithData)
}
