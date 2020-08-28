package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RiderJump is sent by the client to the server when it jumps while riding an entity that has the
// WASDControlled entity flag set, for example when riding a horse.
type RiderJump struct {
	// JumpStrength is the strength of the jump, depending on how long the rider has held the jump button.
	JumpStrength int32
}

// ID ...
func (*RiderJump) ID() uint32 {
	return IDRiderJump
}

// Marshal ...
func (pk *RiderJump) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.JumpStrength)
}

// Unmarshal ...
func (pk *RiderJump) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.JumpStrength)
}
