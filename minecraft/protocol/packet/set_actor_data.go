package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetActorData is sent by the server to update the entity metadata of an entity. It includes flags such as
// if the entity is on fire, but also properties such as the air it has left until it starts drowning.
type SetActorData struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks. Flags include ones such as 'on fire' and 'sprinting'.
	// The metadata values are indexed by their property key.
	EntityMetadata map[uint32]any
	// EntityProperties is a list of properties that the entity inhibits. These properties define and alter specific
	// attributes of the entity.
	EntityProperties protocol.EntityProperties
	// Tick is the server tick at which the packet was sent. It is used in relation to CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*SetActorData) ID() uint32 {
	return IDSetActorData
}

func (pk *SetActorData) Marshal(io protocol.IO) {
	io.Varuint64(&pk.EntityRuntimeID)
	io.EntityMetadata(&pk.EntityMetadata)
	protocol.Single(io, &pk.EntityProperties)
	io.Varuint64(&pk.Tick)
}
