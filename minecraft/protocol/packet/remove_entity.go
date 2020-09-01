package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveEntity is sent by the server to the client. Its function is not entirely clear: It does not remove an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type RemoveEntity struct {
	// EntityNetworkID is the network ID of the entity that should be removed.
	EntityNetworkID uint64
}

// ID ...
func (pk *RemoveEntity) ID() uint32 {
	return IDRemoveEntity
}

// Marshal ...
func (pk *RemoveEntity) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityNetworkID)
}

// Unmarshal ...
func (pk *RemoveEntity) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityNetworkID)
}
