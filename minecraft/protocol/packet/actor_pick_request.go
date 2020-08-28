package packet

import (
	"bytes"
	"encoding/binary"
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
}

// ID ...
func (*ActorPickRequest) ID() uint32 {
	return IDActorPickRequest
}

// Marshal ...
func (pk *ActorPickRequest) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.EntityUniqueID)
	_ = binary.Write(buf, binary.LittleEndian, pk.HotBarSlot)
}

// Unmarshal ...
func (pk *ActorPickRequest) Unmarshal(r *protocol.Reader) {
	r.Int64(&pk.EntityUniqueID)
	r.Uint8(&pk.HotBarSlot)
}
