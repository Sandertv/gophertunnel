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
	// Tick is the server tick at which the packet was sent. It is used in relation to CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*SetActorData) ID() uint32 {
	return IDSetActorData
}

// Marshal ...
func (pk *SetActorData) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.EntityMetadata(&pk.EntityMetadata)
	w.Varuint64(&pk.Tick)
}

// Unmarshal ...
func (pk *SetActorData) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.EntityMetadata(&pk.EntityMetadata)
	r.Varuint64(&pk.Tick)
}
