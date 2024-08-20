package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerArmourDamageFlagHelmet = 1 << iota
	PlayerArmourDamageFlagChestplate
	PlayerArmourDamageFlagLeggings
	PlayerArmourDamageFlagBoots
	PlayerArmourDamageFlagBody
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
	// BodyDamage is the amount of damage that should be dealt to the body.
	BodyDamage int32
}

// ID ...
func (pk *PlayerArmourDamage) ID() uint32 {
	return IDPlayerArmourDamage
}

func (pk *PlayerArmourDamage) Marshal(io protocol.IO) {
	io.Uint8(&pk.Bitset)
	if pk.Bitset&PlayerArmourDamageFlagHelmet != 0 {
		io.Varint32(&pk.HelmetDamage)
	} else {
		pk.HelmetDamage = 0
	}
	if pk.Bitset&PlayerArmourDamageFlagChestplate != 0 {
		io.Varint32(&pk.ChestplateDamage)
	} else {
		pk.ChestplateDamage = 0
	}
	if pk.Bitset&PlayerArmourDamageFlagLeggings != 0 {
		io.Varint32(&pk.LeggingsDamage)
	} else {
		pk.LeggingsDamage = 0
	}
	if pk.Bitset&PlayerArmourDamageFlagBoots != 0 {
		io.Varint32(&pk.BootsDamage)
	} else {
		pk.BootsDamage = 0
	}
	if pk.Bitset&PlayerArmourDamageFlagBody != 0 {
		io.Varint32(&pk.BodyDamage)
	} else {
		pk.BodyDamage = 0
	}
}
