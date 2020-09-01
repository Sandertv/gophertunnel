package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerToClientHandshake is sent by the server to the client to complete the key exchange in order to
// initialise encryption on client and server side. It is followed up by a ClientToServerHandshake packet
// from the client.
type ServerToClientHandshake struct {
	// JWT is a raw JWT token containing data such as the public key from the server, the algorithm used and
	// the server's token. It is used for the client to produce a shared secret.
	JWT []byte
}

// ID ...
func (*ServerToClientHandshake) ID() uint32 {
	return IDServerToClientHandshake
}

// Marshal ...
func (pk *ServerToClientHandshake) Marshal(w *protocol.Writer) {
	w.ByteSlice(&pk.JWT)
}

// Unmarshal ...
func (pk *ServerToClientHandshake) Unmarshal(r *protocol.Reader) {
	r.ByteSlice(&pk.JWT)
}
