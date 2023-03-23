package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AvailableCommands is sent by the server to send a list of all commands that the player is able to use on
// the server. This packet holds all the arguments of each commands as well, making it possible for the client
// to provide auto-completion and command usages.
type AvailableCommands struct {
	// Commands is a list of all commands that the client should show client-side. The AvailableCommands
	// packet replaces any commands sent before. It does not only add the commands that are sent in it.
	Commands []protocol.Command
	// Constraints is a list of constraints that should be applied to certain options of enums in the commands
	// above.
	Constraints []protocol.CommandEnumConstraint
}

// ID ...
func (*AvailableCommands) ID() uint32 {
	return IDAvailableCommands
}

func (pk *AvailableCommands) Marshal(io protocol.IO) {
	io.Commands(&pk.Commands, &pk.Constraints)
}
