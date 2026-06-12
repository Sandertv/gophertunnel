package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PartyChanged is sent by the client to the server to indicate that the player's party ID has changed.
type PartyChanged struct {
	PartyInfo protocol.Optional[PartyInfo]
}

// ID ...
func (*PartyChanged) ID() uint32 {
	return IDPartyChanged
}

func (pk *PartyChanged) Marshal(io protocol.IO) {
	protocol.OptionalMarshaler(io, &pk.PartyInfo)
}

// PartyInfo represents the information of the client's role in a party.
type PartyInfo struct {
	// PartyID is the party identifier.
	PartyID string
	// PartyLeader is if the client is the new party leader or not.
	PartyLeader bool
}

func (x *PartyInfo) Marshal(io protocol.IO) {
	io.String(&x.PartyID)
	io.Bool(&x.PartyLeader)
}
