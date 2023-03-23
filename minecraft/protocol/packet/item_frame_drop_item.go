package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemFrameDropItem is sent by the client when it takes an item out of an item frame.
type ItemFrameDropItem struct {
	// Position is the position of the item frame that had its item dropped. There must be a 'block entity'
	// present at this position.
	Position protocol.BlockPos
}

// ID ...
func (*ItemFrameDropItem) ID() uint32 {
	return IDItemFrameDropItem
}

func (pk *ItemFrameDropItem) Marshal(io protocol.IO) {
	io.UBlockPos(&pk.Position)
}
