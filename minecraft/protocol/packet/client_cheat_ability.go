package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCheatAbility ...
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
	protocol.Single(w, &pk.AbilityData)
}

// Unmarshal ...
func (pk *ClientCheatAbility) Unmarshal(r *protocol.Reader) {
	protocol.Single(r, &pk.AbilityData)
}
