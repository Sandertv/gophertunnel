package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// InventorySlot is sent by the server to update a single slot in one of the inventory windows that the client
// currently has opened. Usually this is the main inventory, but it may also be the off hand or, for example,
// a chest inventory.
type InventorySlot struct {
	// WindowID is the ID of the window that the packet modifies. It must point to one of the windows that the
	// client currently has opened.
	WindowID uint32
	// Slot is the index of the slot that the packet modifies. The new item will be set to the slot at this
	// index.
	Slot uint32
	// DynamicWindowID is the ID of the window if it is dynamic. If the window is not dynamic, this field is
	// set to 0.
	DynamicWindowID uint32
	// NewItem is the item to be put in the slot at Slot. It will overwrite any item that may currently
	// be present in that slot.
	NewItem protocol.ItemInstance
}

// ID ...
func (*InventorySlot) ID() uint32 {
	return IDInventorySlot
}

func (pk *InventorySlot) Marshal(io protocol.IO) {
	io.Varuint32(&pk.WindowID)
	io.Varuint32(&pk.Slot)
	io.Varuint32(&pk.DynamicWindowID)
	io.ItemInstance(&pk.NewItem)
}
