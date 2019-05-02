package packet

import "bytes"

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
func (pk *Unknown) Marshal(buf *bytes.Buffer) {
	_, _ = buf.Write(pk.Payload)
}

// Unmarshal ...
func (pk *Unknown) Unmarshal(buf *bytes.Buffer) error {
	pk.Payload = buf.Bytes()
	buf.Reset()
	return nil
}
