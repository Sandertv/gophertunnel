package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
func (pk *SetCommandsEnabled) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SetCommandsEnabled) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SetCommandsEnabled) marshal(r protocol.IO) {
	r.Bool(&pk.Enabled)
}
