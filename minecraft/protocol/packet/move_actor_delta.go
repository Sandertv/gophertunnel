package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	MoveActorDeltaFlagHasX = 1 << iota
	MoveActorDeltaFlagHasY
	MoveActorDeltaFlagHasZ
	MoveActorDeltaFlagHasRotX
	MoveActorDeltaFlagHasRotY
	MoveActorDeltaFlagHasRotZ
	MoveActorDeltaFlagOnGround
	MoveActorDeltaFlagTeleport
	MoveActorDeltaFlagForceMove
)

// MoveActorDelta is sent by the server to move an entity. The packet is specifically optimised to save as
// much space as possible, by only writing non-zero fields.
// As of 1.16.100, this packet no longer actually contains any deltas.
type MoveActorDelta struct {
	// Flags is a list of flags that specify what data is in the packet.
	Flags uint16
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a
	// non-player entity with this runtime ID is present.
	EntityRuntimeID uint64
	// Position is the new position that the entity was moved to.
	Position mgl32.Vec3
	// Rotation is the new absolute rotation. Unlike the position, it is not actually a delta. If any of the
	// values of this rotation are not sent, these values are 0 and no flag for them is present.
	Rotation mgl32.Vec3
}

// ID ...
func (*MoveActorDelta) ID() uint32 {
	return IDMoveActorDelta
}

// Marshal ...
func (pk *MoveActorDelta) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint16(&pk.Flags)
	if pk.Flags&MoveActorDeltaFlagHasX != 0 {
		w.Float32(&pk.Position[0])
	}
	if pk.Flags&MoveActorDeltaFlagHasY != 0 {
		w.Float32(&pk.Position[1])
	}
	if pk.Flags&MoveActorDeltaFlagHasZ != 0 {
		w.Float32(&pk.Position[2])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotX != 0 {
		w.ByteFloat(&pk.Rotation[0])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotY != 0 {
		w.ByteFloat(&pk.Rotation[1])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotZ != 0 {
		w.ByteFloat(&pk.Rotation[2])
	}
}

// Unmarshal ...
func (pk *MoveActorDelta) Unmarshal(r *protocol.Reader) {
	pk.Position = mgl32.Vec3{}
	pk.Rotation = mgl32.Vec3{}
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint16(&pk.Flags)
	if pk.Flags&MoveActorDeltaFlagHasX != 0 {
		r.Float32(&pk.Position[0])
	}
	if pk.Flags&MoveActorDeltaFlagHasY != 0 {
		r.Float32(&pk.Position[1])
	}
	if pk.Flags&MoveActorDeltaFlagHasZ != 0 {
		r.Float32(&pk.Position[2])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotX != 0 {
		r.ByteFloat(&pk.Rotation[0])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotY != 0 {
		r.ByteFloat(&pk.Rotation[1])
	}
	if pk.Flags&MoveActorDeltaFlagHasRotZ != 0 {
		r.ByteFloat(&pk.Rotation[2])
	}
}
