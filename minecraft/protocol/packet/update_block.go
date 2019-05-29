package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BlockUpdateNeighbours = 1 << iota
	BlockUpdateNetwork
	BlockUpdateNoGraphics
	BlockUpdatePriority
)

// UpdateBlock is sent by the server to update a block client-side, without resending the entire chunk that
// the block is located in. It is particularly useful for small modifications like block breaking/placing.
type UpdateBlock struct {
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
}

// ID ...
func (*UpdateBlock) ID() uint32 {
	return IDUpdateBlock
}

// Marshal ...
func (pk *UpdateBlock) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = protocol.WriteVaruint32(buf, pk.NewBlockRuntimeID)
	_ = protocol.WriteVaruint32(buf, pk.Flags)
	_ = protocol.WriteVaruint32(buf, pk.Layer)
}

// Unmarshal ...
func (pk *UpdateBlock) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.UBlockPosition(buf, &pk.Position),
		protocol.Varuint32(buf, &pk.NewBlockRuntimeID),
		protocol.Varuint32(buf, &pk.Flags),
		protocol.Varuint32(buf, &pk.Layer),
	)
}
