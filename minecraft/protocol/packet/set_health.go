package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetHealth is sent by the server. It sets the health of the player it is sent to. The SetHealth packet
// should no longer be used. Instead, the health attribute should be used so that the health and maximum
// health may be changed directly.
type SetHealth struct {
	// Health is the new health of the player.
	Health int32
}

// ID ...
func (*SetHealth) ID() uint32 {
	return IDSetHealth
}

func (pk *SetHealth) Marshal(io protocol.IO) {
	io.Varint32(&pk.Health)
}
