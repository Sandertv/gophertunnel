package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MobArmourEquipment is sent by the server to the client to update the armour an entity is wearing. It is
// sent for both players and other entities, such as zombies.
type MobArmourEquipment struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Helmet is the equipped helmet of the entity. Items that are not wearable on the head will not be
	// rendered by the client. Unlike in Java Edition, blocks cannot be worn.
	Helmet protocol.ItemInstance
	// Chestplate is the chestplate of the entity. Items that are not wearable as chestplate will not be
	// rendered.
	Chestplate protocol.ItemInstance
	// Leggings is the item worn as leggings by the entity. Items not wearable as leggings will not be
	// rendered client-side.
	Leggings protocol.ItemInstance
	// Boots is the item worn as boots by the entity. Items not wearable as boots will not be rendered.
	Boots protocol.ItemInstance
}

// ID ...
func (*MobArmourEquipment) ID() uint32 {
	return IDMobArmourEquipment
}

// Marshal ...
func (pk *MobArmourEquipment) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.ItemInstance(&pk.Helmet)
	w.ItemInstance(&pk.Chestplate)
	w.ItemInstance(&pk.Leggings)
	w.ItemInstance(&pk.Boots)
}

// Unmarshal ...
func (pk *MobArmourEquipment) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.ItemInstance(&pk.Helmet)
	r.ItemInstance(&pk.Chestplate)
	r.ItemInstance(&pk.Leggings)
	r.ItemInstance(&pk.Boots)
}
