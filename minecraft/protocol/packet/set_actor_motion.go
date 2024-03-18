package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetActorMotion is sent by the server to change the client-side velocity of an entity. It is usually used
// in combination with server-side movement calculation.
type SetActorMotion struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Velocity is the new velocity the entity gets. This velocity will initiate the client-side movement of
	// the entity.
	Velocity mgl32.Vec3
	// Tick is the server tick at which the packet was sent. It is used in relation to CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*SetActorMotion) ID() uint32 {
	return IDSetActorMotion
}

func (pk *SetActorMotion) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.Vec3(&pk.Velocity)
	io.Varuint64(&pk.Tick)
}
