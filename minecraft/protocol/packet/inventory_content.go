package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// InventoryContent is sent by the server to update the full content of a particular inventory. It is usually
// sent for the main inventory of the player, but also works for other inventories that are currently opened
// by the player.
type InventoryContent struct {
	// WindowID is the ID that identifies one of the windows that the client currently has opened, or one of
	// the consistent windows such as the main inventory.
	WindowID uint32
	// Content is the new content of the inventory. The length of this slice must be equal to the full size of
	// the inventory window updated.
	Content []protocol.ItemInstance
	// Container is the protocol.FullContainerName that describes the container that the content is for.
	Container protocol.FullContainerName
	// DynamicContainerSize is the size of the container, if the container is dynamic.
	DynamicContainerSize uint32
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return IDInventoryContent
}

func (pk *InventoryContent) Marshal(io protocol.IO) {
	io.Varuint32(&pk.WindowID)
	protocol.FuncSlice(io, &pk.Content, io.ItemInstance)
	protocol.Single(io, &pk.Container)
	io.Varuint32(&pk.DynamicContainerSize)
}
