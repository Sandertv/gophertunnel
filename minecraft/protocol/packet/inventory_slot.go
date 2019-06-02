package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// InventorySlot is sent by the server to update a single slot in one of the inventory windows that the client
// currently has opened. Usually this is the main inventory, but it may also be the off hand or, for example,
// a chest inventory.
type InventorySlot struct {
	// WindowID is the ID of the window that the packet modifies. It must point to one of the windows that the
	// client currently has opened.
	WindowID uint32
	// SlotIndex is the index of the slot that the packet modifies. The new item will be set to the slot at
	// this index.
	SlotIndex uint32
	// NewItem is the item to be put in the slot at SlotIndex. It will overwrite any item that may currently
	// be present in that slot.
	NewItem protocol.ItemStack
}

// ID ...
func (*InventorySlot) ID() uint32 {
	return IDInventorySlot
}

// Marshal ...
func (pk *InventorySlot) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint32(buf, pk.WindowID)
	_ = protocol.WriteVaruint32(buf, pk.SlotIndex)
	_ = protocol.WriteItem(buf, pk.NewItem)
}

// Unmarshal ...
func (pk *InventorySlot) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint32(buf, &pk.WindowID),
		protocol.Varuint32(buf, &pk.SlotIndex),
		protocol.Item(buf, &pk.NewItem),
	)
}
