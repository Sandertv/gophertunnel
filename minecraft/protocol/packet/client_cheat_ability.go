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

// Marshal ...
func (pk *ClientCheatAbility) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ClientCheatAbility) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ClientCheatAbility) marshal(r protocol.IO) {
	protocol.Single(r, &pk.AbilityData)
}
