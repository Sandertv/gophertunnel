package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
)

// Packet represents a packet that may be sent over a Minecraft network connection. The packet needs to hold
// a method to encode itself to binary and decode itself from binary.
type Packet interface {
	// ID returns the ID of the packet. All of these identifiers of packets may be found in id.go.
	ID() uint32
	// Marshal encodes or decodes a Packet, depending on the protocol.IO
	// implementation passed. When passing a protocol.Writer, Marshal will
	// encode the Packet into its binary representation and write it to the
	// protocol.Writer. On the other hand, when passing a protocol.Reader,
	// Marshal will decode the bytes from the reader into the Packet.
	Marshal(io protocol.IO)
}

// Header is the header of a packet. It exists out of a single varuint32 which is composed of a packet ID and
// a sender and target sub client ID. These IDs are used for split screen functionality.
type Header struct {
	PacketID        uint32
	SenderSubClient byte
	TargetSubClient byte
}

// Write writes the header as a single varuint32 to buf.
func (header *Header) Write(w io.ByteWriter) error {
	return protocol.WriteVaruint32(w, header.PacketID|(uint32(header.SenderSubClient)<<10)|(uint32(header.TargetSubClient)<<12))
}

// Read reads a varuint32 from buf and sets the corresponding values to the Header.
func (header *Header) Read(r io.ByteReader) error {
	var value uint32
	if err := protocol.Varuint32(r, &value); err != nil {
		return err
	}
	header.PacketID = value & 0x3FF
	header.SenderSubClient = byte((value >> 10) & 0x3)
	header.TargetSubClient = byte((value >> 12) & 0x3)
	return nil
}
