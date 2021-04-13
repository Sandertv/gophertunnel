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
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return IDInventoryContent
}

// Marshal ...
func (pk *InventoryContent) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Content))
	w.Varuint32(&pk.WindowID)
	w.Varuint32(&l)
	for _, item := range pk.Content {
		w.ItemInstance(&item)
	}
}

// Unmarshal ...
func (pk *InventoryContent) Unmarshal(r *protocol.Reader) {
	var length uint32
	r.Varuint32(&pk.WindowID)
	r.Varuint32(&length)

	pk.Content = make([]protocol.ItemInstance, length)
	for i := uint32(0); i < length; i++ {
		r.ItemInstance(&pk.Content[i])
	}
}
