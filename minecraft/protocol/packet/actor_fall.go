package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ActorFall is sent by the client when it falls from a distance onto a block that would damage the player.
// This packet should not be used at all by the server, as it can easily be spoofed using a proxy or custom
// client. Servers should implement fall damage using their own calculations.
type ActorFall struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// FallDistance is the distance that the entity fell until it hit the ground. The damage would otherwise
	// be calculated using this field.
	FallDistance float32
	// InVoid specifies if the fall was in the void. The player can't fall below roughly Y=-40.
	InVoid bool
}

// ID ...
func (*ActorFall) ID() uint32 {
	return IDActorFall
}

// Marshal ...
func (pk *ActorFall) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Float32(&pk.FallDistance)
	w.Bool(&pk.InVoid)
}

// Unmarshal ...
func (pk *ActorFall) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Float32(&pk.FallDistance)
	r.Bool(&pk.InVoid)
}
