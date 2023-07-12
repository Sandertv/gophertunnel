package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AgentAnimation is an Education Edition packet sent from the server to the client to make an agent perform an animation.
type AgentAnimation struct {
	// Animation is the ID of the animation that the agent should perform. As of its implementation, there are no IDs
	// that can be used in the regular client.
	Animation byte
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
}

// ID ...
func (*AgentAnimation) ID() uint32 {
	return IDAgentAnimation
}

func (pk *AgentAnimation) Marshal(io protocol.IO) {
	io.Uint8(&pk.Animation)
	io.Varuint64(&pk.EntityRuntimeID)
}
