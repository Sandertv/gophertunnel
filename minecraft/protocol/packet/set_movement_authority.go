package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetMovementAuthority is sent by the server to the client to change its movement mode.
type SetMovementAuthority struct {
	// MovementType specifies the way the server handles player movement. Available options are
	// protocol.PlayerMovementModeClient, protocol.PlayerMovementModeServer and
	// protocol.PlayerMovementModeServerWithRewind, where the server authoritative types result
	// in the client sending PlayerAuthInput packets instead of MovePlayer packets and the rewind mode
	// requires sending the tick of movement and several actions.
	MovementType byte
}

// ID ...
func (*SetMovementAuthority) ID() uint32 {
	return IDSetMovementAuthority
}

func (pk *SetMovementAuthority) Marshal(io protocol.IO) {
	io.Uint8(&pk.MovementType)
}
