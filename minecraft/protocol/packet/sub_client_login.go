package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubClientLogin is sent when a sub-client joins the server while another client is already connected to it.
// The packet is sent as a result of split-screen game play, and allows up to four players to play using the
// same network connection. After an initial Login packet from the 'main' client, each sub-client that
// connects sends a SubClientLogin to request their own login.
type SubClientLogin struct {
	// ConnectionRequest is a string containing information about the player and JWTs that may be used to
	// verify if the player is connected to XBOX Live. The connection request also contains the necessary
	// client public key to initiate encryption.
	// The ConnectionRequest in this packet is identical to the one found in the Login packet.
	ConnectionRequest []byte
}

// ID ...
func (*SubClientLogin) ID() uint32 {
	return IDSubClientLogin
}

// Marshal ...
func (pk *SubClientLogin) Marshal(w *protocol.Writer) {
	w.ByteSlice(&pk.ConnectionRequest)
}

// Unmarshal ...
func (pk *SubClientLogin) Unmarshal(r *protocol.Reader) {
	r.ByteSlice(&pk.ConnectionRequest)
}
