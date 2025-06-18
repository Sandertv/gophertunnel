package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientMovementPredictionSync is sent by the client to the server periodically if the client has received
// movement corrections from the server, containing information about client-predictions that are relevant
// to movement.
type ClientMovementPredictionSync struct {
	// ActorFlags is a bitset of all the flags that are currently set for the client.
	ActorFlags protocol.Bitset
	// BoundingBoxScale is the scale of the client's bounding box.
	BoundingBoxScale float32
	// BoundingBoxWidth is the width of the client's bounding box.
	BoundingBoxWidth float32
	// BoundingBoxHeight is the height of the client's bounding box.
	BoundingBoxHeight float32
	// MovementSpeed is the movement speed attribute or 0 if not set.
	MovementSpeed float32
	// UnderwaterMovementSpeed is the underwater movement speed attribute or 0 if not set.
	UnderwaterMovementSpeed float32
	// LavaMovementSpeed is the lava movement speed attribute or 0 if not set.
	LavaMovementSpeed float32
	// JumpStrength is the jump strength attribute or 0 if not set.
	JumpStrength float32
	// Health is the health attribute or 0 if not set.
	Health float32
	// Hunger is the hunger attribute or 0 if not set.
	Hunger float32
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world.
	EntityUniqueID int64
	// Flying specifies if the client is currently flying.
	Flying bool
}

// ID ...
func (*ClientMovementPredictionSync) ID() uint32 {
	return IDClientMovementPredictionSync
}

func (pk *ClientMovementPredictionSync) Marshal(io protocol.IO) {
	io.Bitset(&pk.ActorFlags, protocol.EntityDataFlagCount)
	io.Float32(&pk.BoundingBoxScale)
	io.Float32(&pk.BoundingBoxWidth)
	io.Float32(&pk.BoundingBoxHeight)
	io.Float32(&pk.MovementSpeed)
	io.Float32(&pk.UnderwaterMovementSpeed)
	io.Float32(&pk.LavaMovementSpeed)
	io.Float32(&pk.JumpStrength)
	io.Float32(&pk.Health)
	io.Float32(&pk.Hunger)
	io.Varint64(&pk.EntityUniqueID)
	io.Bool(&pk.Flying)
}
