package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerArmourDamageFlagHelmet = iota
	PlayerArmourDamageFlagChestplate
	PlayerArmourDamageFlagLeggings
	PlayerArmourDamageFlagBoots
	PlayerArmourDamageFlagBody
)

// PlayerArmourDamage is sent by the server to damage the armour of a player. It is a very efficient packet,
// but generally it's much easier to just send a slot update for the damaged armour.
type PlayerArmourDamage struct {
	// List is a list of armour entries indicating which pieces of armour should receive damage.
	List []protocol.PlayerArmourDamageEntry
}

// ID ...
func (pk *PlayerArmourDamage) ID() uint32 {
	return IDPlayerArmourDamage
}

func (pk *PlayerArmourDamage) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.List)
}
