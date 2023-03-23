package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetLocalPlayerAsInitialised is sent by the client in response to a PlayStatus packet with the status set
// to 3. The packet marks the moment at which the client is fully initialised and can receive any packet
// without discarding it.
type SetLocalPlayerAsInitialised struct {
	// EntityRuntimeID is the entity runtime ID the player was assigned earlier in the login sequence in the
	// StartGame packet.
	EntityRuntimeID uint64
}

// ID ...
func (*SetLocalPlayerAsInitialised) ID() uint32 {
	return IDSetLocalPlayerAsInitialised
}

func (pk *SetLocalPlayerAsInitialised) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
}
