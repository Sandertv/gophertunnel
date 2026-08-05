package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MoveActorDelta is sent by the server to move an entity. The packet is specifically optimised to save as
// much space as possible, by only writing non-zero fields.
// As of 1.16.100, this packet no longer actually contains any deltas.
type MoveActorDelta struct {
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a
	// non-player entity with this runtime ID is present.
	EntityRuntimeID uint64
	// Position is the new position that the entity was moved to.
	PositionX protocol.Optional[float32]
	PositionY protocol.Optional[float32]
	PositionZ protocol.Optional[float32]
	// Rotation is the new absolute rotation. Unlike the position, it is not actually a delta. If any of the
	// values of this rotation are not sent, these values are 0 and no flag for them is present.
	RotationX     protocol.Optional[float32]
	RotationY     protocol.Optional[float32]
	RotationYHead protocol.Optional[float32]

	OnGround             bool
	ForceMove            bool
	ForceMoveLocalEntity bool
	ForceCompletion      bool
}

// ID ...
func (*MoveActorDelta) ID() uint32 {
	return IDMoveActorDelta
}

func (pk *MoveActorDelta) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	protocol.OptionalFunc(io, &pk.PositionX, io.Float32)
	protocol.OptionalFunc(io, &pk.PositionY, io.Float32)
	protocol.OptionalFunc(io, &pk.PositionZ, io.Float32)
	protocol.OptionalFunc(io, &pk.RotationX, io.ByteFloat)
	protocol.OptionalFunc(io, &pk.RotationY, io.ByteFloat)
	protocol.OptionalFunc(io, &pk.RotationYHead, io.ByteFloat)
	io.Bool(&pk.OnGround)
	io.Bool(&pk.ForceMove)
	io.Bool(&pk.ForceMoveLocalEntity)
	io.Bool(&pk.ForceCompletion)
}
