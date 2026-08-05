package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MoveActorDelta is sent by the server to move an entity. Each position and rotation component has an independent
// presence flag so unchanged components may be omitted.
type MoveActorDelta struct {
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a
	// non-player entity with this runtime ID is present.
	EntityRuntimeID uint64
	// PositionX, PositionY and PositionZ are the optional components of the entity's new absolute position.
	PositionX protocol.Optional[float32]
	PositionY protocol.Optional[float32]
	PositionZ protocol.Optional[float32]
	// RotationX, RotationY and RotationYHead are the optional components of the entity's new absolute rotation.
	RotationX     protocol.Optional[float32]
	RotationY     protocol.Optional[float32]
	RotationYHead protocol.Optional[float32]

	// OnGround specifies whether the entity is on the ground after the movement.
	OnGround bool
	// ForceMove forces the movement to be applied.
	ForceMove bool
	// ForceMoveLocalEntity applies the forced movement to the local player entity.
	ForceMoveLocalEntity bool
	// ForceCompletion forces the movement to complete immediately.
	ForceCompletion bool
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
