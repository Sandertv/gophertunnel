package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PassengerJump is sent by the client to the server when it jumps while riding an entity that has the
// WASDControlled entity flag set, for example when riding a horse.
type PassengerJump struct {
	// JumpStrength is the strength of the jump, depending on how long the rider has held the jump button.
	JumpStrength int32
}

// ID ...
func (*PassengerJump) ID() uint32 {
	return IDPassengerJump
}

func (pk *PassengerJump) Marshal(io protocol.IO) {
	io.Varint32(&pk.JumpStrength)
}
