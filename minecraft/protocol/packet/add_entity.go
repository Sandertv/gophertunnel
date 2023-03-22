package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddEntity is sent by the server to the client. Its function is not entirely clear: It does not add an
// entity in the sense of an in-game entity, but has to do with the ECS that Minecraft uses.
type AddEntity struct {
	// EntityNetworkID is the network ID of the entity that should be added.
	EntityNetworkID uint64
}

// ID ...
func (pk *AddEntity) ID() uint32 {
	return IDAddEntity
}

// Marshal ...
func (pk *AddEntity) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *AddEntity) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *AddEntity) marshal(r protocol.IO) {
	r.Varuint64(&pk.EntityNetworkID)
}
