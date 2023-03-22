package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ShowProfile is sent by the server to show the XBOX Live profile of one player to another.
type ShowProfile struct {
	// XUID is the XBOX Live User ID of the player whose profile should be shown to the player. If it is not
	// a valid XUID, the client ignores the packet.
	XUID string
}

// ID ...
func (*ShowProfile) ID() uint32 {
	return IDShowProfile
}

// Marshal ...
func (pk *ShowProfile) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ShowProfile) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ShowProfile) marshal(r protocol.IO) {
	r.String(&pk.XUID)
}
