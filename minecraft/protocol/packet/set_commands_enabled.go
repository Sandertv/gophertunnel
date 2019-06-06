package packet

import (
	"bytes"
	"encoding/binary"
)

// SetCommandsEnabled is sent by the server to enable or disable the ability to execute commands for the
// client. If disabled, the client itself will stop the execution of commands.
type SetCommandsEnabled struct {
	// Enabled defines if the commands should be enabled, or if false, disabled.
	Enabled bool
}

// ID ...
func (*SetCommandsEnabled) ID() uint32 {
	return IDSetCommandsEnabled
}

// Marshal ...
func (pk *SetCommandsEnabled) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Enabled)
}

// Unmarshal ...
func (pk *SetCommandsEnabled) Unmarshal(buf *bytes.Buffer) error {
	return binary.Read(buf, binary.LittleEndian, &pk.Enabled)
}
