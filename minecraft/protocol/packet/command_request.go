package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CommandRequest is sent by the client to request the execution of a server-side command. Although some
// servers support sending commands using the Text packet, this packet is guaranteed to have the correct
// result.
type CommandRequest struct {
	// CommandLine is the raw entered command line. The client does no parsing of the command line by itself
	// (unlike it did in the early stages), but lets the server do that.
	CommandLine string
	// CommandOrigin is the data specifying the origin of the command. In other words, the source that the
	// command was from, such as the player itself or a websocket server.
	CommandOrigin protocol.CommandOrigin
	// Internal specifies if the command request internal. Setting it to false seems to work and the usage of
	// this field is not known.
	Internal bool
}

// ID ...
func (*CommandRequest) ID() uint32 {
	return IDCommandRequest
}

// Marshal ...
func (pk *CommandRequest) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.CommandLine)
	_ = protocol.WriteCommandOriginData(buf, pk.CommandOrigin)
	_ = binary.Write(buf, binary.LittleEndian, pk.Internal)
}

// Unmarshal ...
func (pk *CommandRequest) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.CommandLine),
		protocol.CommandOriginData(buf, &pk.CommandOrigin),
		binary.Read(buf, binary.LittleEndian, &pk.Internal),
	)
}
