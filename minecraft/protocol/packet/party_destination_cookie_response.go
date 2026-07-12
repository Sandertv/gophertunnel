package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// PartyDestinationCookieResponse is sent by the client to the server in response to a
// SendPartyDestinationCookie packet.
type PartyDestinationCookieResponse struct {
	// Cookie is the opaque party destination cookie echoed back from the SendPartyDestinationCookie packet.
	Cookie string
	// Accepted is true if the client accepted the party destination.
	Accepted bool
}

// ID ...
func (*PartyDestinationCookieResponse) ID() uint32 {
	return IDPartyDestinationCookieResponse
}

func (pk *PartyDestinationCookieResponse) Marshal(io protocol.IO) {
	io.String(&pk.Cookie)
	io.Bool(&pk.Accepted)
}
