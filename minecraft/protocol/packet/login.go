package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Login is sent when the client initially tries to join the server. It is the first packet sent and contains
// information specific to the player.
type Login struct {
	// ClientProtocol is the protocol version of the player. The player is disconnected if the protocol is
	// incompatible with the protocol of the server.
	ClientProtocol int32
	// ConnectionRequest is a string containing information about the player and JWTs that may be used to
	// verify if the player is connected to XBOX Live. The connection request also contains the necessary
	// client public key to initiate encryption.
	ConnectionRequest string
}

// ID ...
func (*Login) ID() uint32 {
	return IDLogin
}

// Marshal ...
func (pk *Login) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.BigEndian, pk.ClientProtocol)
	_ = protocol.WriteString(buf, pk.ConnectionRequest)
}

// Unmarshal ...
func (pk *Login) Unmarshal(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &pk.ClientProtocol); err != nil {
		return err
	}
	if err := protocol.String(buf, &pk.ConnectionRequest); err != nil {
		return err
	}
	return nil
}
