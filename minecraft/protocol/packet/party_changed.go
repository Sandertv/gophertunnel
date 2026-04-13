package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PartyChanged is sent by the client to the server to indicate that the player's party ID has changed.
type PartyChanged struct {
	// PartyID is the party identifier.
	PartyID string
}

// ID ...
func (*PartyChanged) ID() uint32 {
	return IDPartyChanged
}

func (pk *PartyChanged) Marshal(io protocol.IO) {
	io.String(&pk.PartyID)
}
