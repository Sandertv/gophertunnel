package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCheatAbility functions the same as UpdateAbilities. It is unclear why these two are separated.
type ClientCheatAbility struct {
	// AbilityData represents various data about the abilities of a player, such as ability layers or permissions.
	AbilityData protocol.AbilityData
}

// ID ...
func (*ClientCheatAbility) ID() uint32 {
	return IDClientCheatAbility
}

func (pk *ClientCheatAbility) Marshal(io protocol.IO) {
	protocol.Single(io, &pk.AbilityData)
}
