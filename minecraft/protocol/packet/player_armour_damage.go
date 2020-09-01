package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerArmourDamage is sent by the server to damage the armour of a player. It is a very efficient packet,
// but generally it's much easier to just send a slot update for the damaged armour.
type PlayerArmourDamage struct {
	// Bitset holds a bitset of 4 bits that indicate which pieces of armour need to have damage dealt to them.
	// The first bit, when toggled, is for a helmet, the second for the chestplate, the third for the leggings
	// and the fourth for boots.
	Bitset uint8
	// HelmetDamage is the amount of damage that should be dealt to the helmet.
	HelmetDamage int32
	// ChestplateDamage is the amount of damage that should be dealt to the chestplate.
	ChestplateDamage int32
	// LeggingsDamage is the amount of damage that should be dealt to the leggings.
	LeggingsDamage int32
	// BootsDamage is the amount of damage that should be dealt to the boots.
	BootsDamage int32
}

// ID ...
func (pk *PlayerArmourDamage) ID() uint32 {
	return IDPlayerArmourDamage
}

// Marshal ...
func (pk *PlayerArmourDamage) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.Bitset)
	if pk.Bitset&0b0001 != 0 {
		w.Varint32(&pk.HelmetDamage)
	}
	if pk.Bitset&0b0010 != 0 {
		w.Varint32(&pk.ChestplateDamage)
	}
	if pk.Bitset&0b0100 != 0 {
		w.Varint32(&pk.LeggingsDamage)
	}
	if pk.Bitset&0b1000 != 0 {
		w.Varint32(&pk.BootsDamage)
	}
}

// Unmarshal ...
func (pk *PlayerArmourDamage) Unmarshal(r *protocol.Reader) {
	pk.HelmetDamage, pk.ChestplateDamage, pk.LeggingsDamage, pk.BootsDamage = 0, 0, 0, 0

	r.Uint8(&pk.Bitset)
	if pk.Bitset&0b0001 != 0 {
		r.Varint32(&pk.HelmetDamage)
	}
	if pk.Bitset&0b0010 != 0 {
		r.Varint32(&pk.ChestplateDamage)
	}
	if pk.Bitset&0b0100 != 0 {
		r.Varint32(&pk.LeggingsDamage)
	}
	if pk.Bitset&0b1000 != 0 {
		r.Varint32(&pk.BootsDamage)
	}
}
