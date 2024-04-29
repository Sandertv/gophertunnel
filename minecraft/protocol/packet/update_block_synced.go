package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BlockToEntityTransition = iota + 1
	EntityToBlockTransition
)

// UpdateBlockSynced is sent by the server to synchronise the falling of a falling block entity with the
// transitioning back and forth from and to a solid block. It is used to prevent the entity from flickering,
// and is used in places such as the pushing of blocks with pistons.
type UpdateBlockSynced struct {
	// Position is the block position at which a block is updated.
	Position protocol.BlockPos
	// NewBlockRuntimeID is the runtime ID of the block that is placed at Position after sending the packet
	// to the client.
	NewBlockRuntimeID uint32
	// Flags is a combination of flags that specify the way the block is updated client-side. It is a
	// combination of the flags above, but typically sending only the BlockUpdateNetwork flag is sufficient.
	Flags uint32
	// Layer is the world layer on which the block is updated. For most blocks, this is the first layer, as
	// that layer is the default layer to place blocks on, but for blocks inside of each other, this differs.
	Layer uint32
	// EntityUniqueID is the unique ID of the falling block entity that the block transitions to or that the
	// entity transitions from.
	// Note that for both possible values for TransitionType, the EntityUniqueID should point to the falling
	// block entity involved.
	EntityUniqueID uint64
	// TransitionType is the type of the transition that happened. It is either BlockToEntityTransition, when
	// a block placed becomes a falling entity, or EntityToBlockTransition, when a falling entity hits the
	// ground and becomes a solid block again.
	TransitionType uint64
}

// ID ...
func (*UpdateBlockSynced) ID() uint32 {
	return IDUpdateBlockSynced
}

func (pk *UpdateBlockSynced) Marshal(io protocol.IO) {
	io.UBlockPos(&pk.Position)
	io.Varuint32(&pk.NewBlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
	io.Varuint64(&pk.EntityUniqueID)
	io.Varuint64(&pk.TransitionType)
}
