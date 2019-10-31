package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SettingsCommand is sent by the client when it sends a command that changes settings. The actual use of this
// packet is unknown.
type SettingsCommand struct {
	// CommandLine is the full command line that the client sent to update a setting.
	CommandLine string
	// SuppressOutput specifies if the client requests the suppressing of the output of the command that was
	// executed.
	SuppressOutput bool
}

// ID ...
func (*SettingsCommand) ID() uint32 {
	return IDSettingsCommand
}

// Marshal ...
func (pk *SettingsCommand) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.CommandLine)
	_ = binary.Write(buf, binary.LittleEndian, pk.SuppressOutput)
}

// Unmarshal ...
func (pk *SettingsCommand) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.CommandLine),
		binary.Read(buf, binary.LittleEndian, &pk.SuppressOutput),
	)
}
