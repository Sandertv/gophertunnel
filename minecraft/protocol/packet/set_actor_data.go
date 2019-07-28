package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetActorData is sent by the server to update the entity metadata of an entity. It includes flags such as
// if the entity is on fire, but also properties such as the air it has left until it starts drowning.
type SetActorData struct {
	// EntityNetworkID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks. Flags include ones such as 'on fire' and 'sprinting'.
	// The metadata values are indexed by their property key.
	EntityMetadata map[uint32]interface{}
}

// ID ...
func (*SetActorData) ID() uint32 {
	return IDSetActorData
}

// Marshal ...
func (pk *SetActorData) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteEntityMetadata(buf, pk.EntityMetadata)
}

// Unmarshal ...
func (pk *SetActorData) Unmarshal(buf *bytes.Buffer) error {
	pk.EntityMetadata = map[uint32]interface{}{}
	return chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.EntityMetadata(buf, &pk.EntityMetadata),
	)
}
