package packet

import (
	"bytes"
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
	// to the client. The runtime ID must point to a block sent in the list in the StartGame packet.
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
	EntityUniqueID int64
	// TransitionType is the type of the transition that happened. It is either BlockToEntityTransition, when
	// a block placed becomes a falling entity, or EntityToBlockTransition, when a falling entity hits the
	// ground and becomes a solid block again.
	TransitionType uint64
}

// ID ...
func (*UpdateBlockSynced) ID() uint32 {
	return IDUpdateBlockSynced
}

// Marshal ...
func (pk *UpdateBlockSynced) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = protocol.WriteVaruint32(buf, pk.NewBlockRuntimeID)
	_ = protocol.WriteVaruint32(buf, pk.Flags)
	_ = protocol.WriteVaruint32(buf, pk.Layer)
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
	_ = protocol.WriteVaruint64(buf, pk.TransitionType)
}

// Unmarshal ...
func (pk *UpdateBlockSynced) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.UBlockPosition(buf, &pk.Position),
		protocol.Varuint32(buf, &pk.NewBlockRuntimeID),
		protocol.Varuint32(buf, &pk.Flags),
		protocol.Varuint32(buf, &pk.Layer),
		protocol.Varint64(buf, &pk.EntityUniqueID),
		protocol.Varuint64(buf, &pk.TransitionType),
	)
}
