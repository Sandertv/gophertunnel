package packet

import (
	"bytes"
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
	Content []protocol.ItemStack
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return IDInventoryContent
}

// Marshal ...
func (pk *InventoryContent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, pk.WindowID)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Content)))
	for _, item := range pk.Content {
		_ = protocol.WriteItem(buf, item)
	}
}

// Unmarshal ...
func (pk *InventoryContent) Unmarshal(buf *bytes.Buffer) error {
	var length uint32
	if err := chainErr(
		protocol.Varuint32(buf, &pk.WindowID),
		protocol.Varuint32(buf, &length),
	); err != nil {
		return err
	}
	pk.Content = make([]protocol.ItemStack, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.Item(buf, &pk.Content[i]); err != nil {
			return err
		}
	}
	return nil
}
