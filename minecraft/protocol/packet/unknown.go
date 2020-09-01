package packet

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Unknown is an implementation of the Packet interface for unknown/unimplemented packets. It holds the packet
// ID and the raw payload. It serves as a way to read raw unknown packets and forward them to another
// connection, without necessarily implementing them.
type Unknown struct {
	// PacketID is the packet ID of the packet.
	PacketID uint32
	// Payload is the raw payload of the packet.
	Payload []byte
}

// ID ...
func (pk *Unknown) ID() uint32 {
	return pk.PacketID
}

// Marshal ...
func (pk *Unknown) Marshal(w *protocol.Writer) {
	w.Bytes(&pk.Payload)
}

// Unmarshal ...
func (pk *Unknown) Unmarshal(r *protocol.Reader) {
	r.Bytes(&pk.Payload)
}

// String implements a hex representation of an unknown packet, so that it is easier to read and identify
// unknown incoming packets.
func (pk *Unknown) String() string {
	return fmt.Sprintf("{ID:0x%x Payload:0x%x}", pk.PacketID, pk.Payload)
}
