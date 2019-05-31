package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Transfer is sent by the server to transfer a player from the current server to another. Doing so will
// fully disconnect the client, bring it back to the main menu and make it connect to the next server.
type Transfer struct {
	// Address is the address of the new server, which might be either a hostname or an actual IP address.
	Address string
	// Port is the UDP port of the new server.
	Port uint16
}

// ID ...
func (*Transfer) ID() uint32 {
	return IDTransfer
}

// Marshal ...
func (pk *Transfer) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.Address)
	_ = binary.Write(buf, binary.LittleEndian, &pk.Port)
}

// Unmarshal ...
func (pk *Transfer) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.Address),
		binary.Read(buf, binary.LittleEndian, &pk.Port),
	)
}
