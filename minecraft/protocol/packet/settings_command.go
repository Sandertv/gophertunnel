package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SettingsCommand is sent by the client when it changes a setting in the settings that results in the issuing
// of a command to the server, such as when Show Coordinates is enabled.
type SettingsCommand struct {
	// CommandLine is the full command line that was sent to the server as a result of the setting that the
	// client changed.
	CommandLine string
	// SuppressOutput specifies if the client requests the suppressing of the output of the command that was
	// executed. Generally this is set to true, as the client won't need a message to confirm the output of
	// the change.
	SuppressOutput bool
}

// ID ...
func (*SettingsCommand) ID() uint32 {
	return IDSettingsCommand
}

// Marshal ...
func (pk *SettingsCommand) Marshal(w *protocol.Writer) {
	w.String(&pk.CommandLine)
	w.Bool(&pk.SuppressOutput)
}

// Unmarshal ...
func (pk *SettingsCommand) Unmarshal(r *protocol.Reader) {
	r.String(&pk.CommandLine)
	r.Bool(&pk.SuppressOutput)
}
