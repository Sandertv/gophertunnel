package packet

import (
	"bytes"
)

// ClientToServerHandshake is sent by the client in response to a ServerToClientHandshake packet sent by the
// server. It is the first encrypted packet in the login handshake and serves as a confirmation that
// encryption is correctly initialised client side.
type ClientToServerHandshake struct {
	// ClientToServerHandshake has no fields.
}

// ID ...
func (*ClientToServerHandshake) ID() uint32 {
	return IDClientToServerHandshake
}

// Marshal ...
func (*ClientToServerHandshake) Marshal(buf *bytes.Buffer) {

}

// Unmarshal ...
func (*ClientToServerHandshake) Unmarshal(buf *bytes.Buffer) error {
	return nil
}
