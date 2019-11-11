package packet

import (
	"bytes"
	"encoding/binary"
)

// TickSync is sent by the client and the server to maintain a synchronized, server-authoritative tick between
// the client and the server. The client sends this packet first, and the server should reply with another one
// of these packets, including the response time.
type TickSync struct {
	// ClientRequestTimestamp is the timestamp on which the client sent this packet to the server. The server
	// should fill out that same value when replying.
	// The ClientRequestTimestamp is always 0.
	ClientRequestTimestamp int64
	// ServerReceptionTimestamp is the timestamp on which the server received the packet sent by the client.
	// When the packet is sent by the client, this value is 0.
	// ServerReceptionTimestamp is generally the current tick of the server. It isn't an actual timestamp, as
	// the field implies.
	ServerReceptionTimestamp int64
}

// ID ...
func (*TickSync) ID() uint32 {
	return IDTickSync
}

// Marshal ...
func (pk *TickSync) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.ClientRequestTimestamp)
	_ = binary.Write(buf, binary.LittleEndian, pk.ServerReceptionTimestamp)
}

// Unmarshal ...
func (pk *TickSync) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.ClientRequestTimestamp),
		binary.Read(buf, binary.LittleEndian, &pk.ServerReceptionTimestamp),
	)
}
