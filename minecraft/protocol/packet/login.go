package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Login is sent when the client initially tries to join the server. It is the first packet sent and contains
// information specific to the player.
type Login struct {
	// ClientProtocol is the protocol version of the player. The player is disconnected if the protocol is incompatible
	// with the protocol of the server. It has been superseded by the protocol version sent in the
	// RequestNetworkSettings packet, so this should no longer be used by the server.
	ClientProtocol int32
	// ConnectionRequest is a string containing information about the player and JWTs that may be used to
	// verify if the player is connected to XBOX Live. The connection request also contains the necessary
	// client public key to initiate encryption.
	ConnectionRequest []byte
}

// ID ...
func (*Login) ID() uint32 {
	return IDLogin
}

func (pk *Login) Marshal(io protocol.IO) {
	io.BEInt32(&pk.ClientProtocol)
	io.ByteSlice(&pk.ConnectionRequest)
}
