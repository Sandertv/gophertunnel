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
	// Container is the protocol.FullContainerName that describes the container that the content is for.
	Container protocol.Optional[protocol.FullContainerName]
	// StorageItem is the item that is acting as the storage container for the inventory. If the inventory is
	// not a dynamic container then this field should be left empty. When set, only the item type is used by
	// the client and none of the other stack info.
	StorageItem protocol.Optional[protocol.ItemInstance]
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
	protocol.OptionalMarshaler(io, &pk.Container)
	protocol.OptionalFunc(io, &pk.StorageItem, io.ItemInstanceNew)
	io.ItemInstanceNew(&pk.NewItem)
}
