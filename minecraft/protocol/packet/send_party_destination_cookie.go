package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	PartyDestinationCookieIntentNotify = "Notify"
	PartyDestinationCookieIntentOptIn  = "OptIn"
	PartyDestinationCookieIntentOptOut = "OptOut"
)

// SendPartyDestinationCookie is sent by the server to a client with a party destination cookie.
type SendPartyDestinationCookie struct {
	// Cookie is the opaque party destination cookie.
	Cookie string
	// Intent is the intent of the cookie. It is one of the PartyDestinationCookieIntent constants.
	Intent string
	// DestinationName is the name of the destination the cookie refers to.
	DestinationName string
}

// ID ...
func (*SendPartyDestinationCookie) ID() uint32 {
	return IDSendPartyDestinationCookie
}

func (pk *SendPartyDestinationCookie) Marshal(io protocol.IO) {
	io.String(&pk.Cookie)
	io.String(&pk.Intent)
	io.String(&pk.DestinationName)
}
