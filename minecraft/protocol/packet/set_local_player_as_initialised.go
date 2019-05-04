package packet

import (
	"bytes"
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

// Marshal ...
func (pk *SetLocalPlayerAsInitialised) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
}

// Unmarshal ...
func (pk *SetLocalPlayerAsInitialised) Unmarshal(buf *bytes.Buffer) error {
	return protocol.Varuint64(buf, &pk.EntityRuntimeID)
}
