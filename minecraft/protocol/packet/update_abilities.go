package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateAbilities is a packet sent from the server to the client to update the abilities of the player. It, along with
// the UpdateAdventureSettings packet, are replacements of the AdventureSettings packet since v1.19.10.
type UpdateAbilities struct {
	// AbilityData represents various data about the abilities of a player, such as ability layers or permissions.
	AbilityData protocol.AbilityData
}

// ID ...
func (*UpdateAbilities) ID() uint32 {
	return IDUpdateAbilities
}

func (pk *UpdateAbilities) Marshal(io protocol.IO) {
	protocol.Single(io, &pk.AbilityData)
}
