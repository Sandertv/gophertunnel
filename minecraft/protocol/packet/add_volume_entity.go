package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddVolumeEntity sends a volume entity's definition and components from server to client.
type AddVolumeEntity struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks.
	EntityMetadata map[string]interface{}
	// EncodingIdentifier ...
	EncodingIdentifier string
	// InstanceIdentifier ...
	InstanceIdentifier string
	// EngineVersion ...
	EngineVersion string
}

// ID ...
func (*AddVolumeEntity) ID() uint32 {
	return IDAddVolumeEntity
}

// Marshal ...
func (pk *AddVolumeEntity) Marshal(w *protocol.Writer) {
	w.Uint64(&pk.EntityRuntimeID)
	w.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	w.String(&pk.EncodingIdentifier)
	w.String(&pk.InstanceIdentifier)
	w.String(&pk.EngineVersion)
}

// Unmarshal ...
func (pk *AddVolumeEntity) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.EntityRuntimeID)
	r.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	r.String(&pk.EncodingIdentifier)
	r.String(&pk.InstanceIdentifier)
	r.String(&pk.EngineVersion)
}
