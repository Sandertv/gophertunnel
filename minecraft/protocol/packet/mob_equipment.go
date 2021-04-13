package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MobEquipment is sent by the client to the server and the server to the client to make the other side
// aware of the new item that an entity is holding. It is used to show the item in the hand of entities such
// as zombies too.
type MobEquipment struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// NewItem is the new item held after sending the MobEquipment packet. The entity will be shown holding
	// that item to the player it was sent to.
	NewItem protocol.ItemInstance
	// InventorySlot is the slot in the inventory that was held. This is the same as HotBarSlot, and only
	// remains for backwards compatibility.
	InventorySlot byte
	// HotBarSlot is the slot in the hot bar that was held. It is the same as InventorySlot, which is only
	// there for backwards compatibility purposes.
	HotBarSlot byte
	// WindowID is the window ID of the window that had its equipped item changed. This is usually the window
	// ID of the normal inventory, but may also be something else, for example with the off hand.
	WindowID byte
}

// ID ...
func (*MobEquipment) ID() uint32 {
	return IDMobEquipment
}

// Marshal ...
func (pk *MobEquipment) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.ItemInstance(&pk.NewItem)
	w.Uint8(&pk.InventorySlot)
	w.Uint8(&pk.HotBarSlot)
	w.Uint8(&pk.WindowID)
}

// Unmarshal ...
func (pk *MobEquipment) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.ItemInstance(&pk.NewItem)
	r.Uint8(&pk.InventorySlot)
	r.Uint8(&pk.HotBarSlot)
	r.Uint8(&pk.WindowID)
}
