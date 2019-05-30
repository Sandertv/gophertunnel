package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// BlockPickRequest is sent by the client when it requests to pick a block in the world and place its item in
// their inventory.
type BlockPickRequest struct {
	// Position is the position at which the client requested to pick the block. The block at that position
	// should have its item put in HotBarSlot if it is empty.
	Position protocol.BlockPos
	// AddBlockNBT specifies if the item should get all NBT tags from the block, meaning the item places a
	// block practically always equal to the one picked.
	AddBlockNBT bool
	// HotBarSlot is the slot that was held at the time of picking a block.
	HotBarSlot byte
}

// ID ...
func (*BlockPickRequest) ID() uint32 {
	return IDBlockPickRequest
}

// Marshal ...
func (pk *BlockPickRequest) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteBlockPosition(buf, pk.Position)
	_ = binary.Write(buf, binary.LittleEndian, pk.AddBlockNBT)
	_ = binary.Write(buf, binary.LittleEndian, pk.HotBarSlot)
}

// Unmarshal ...
func (pk *BlockPickRequest) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.BlockPosition(buf, &pk.Position),
		binary.Read(buf, binary.LittleEndian, &pk.AddBlockNBT),
		binary.Read(buf, binary.LittleEndian, &pk.HotBarSlot),
	)
}
