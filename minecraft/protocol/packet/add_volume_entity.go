package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddVolumeEntity sends a volume entity's definition and metadata from server to client.
type AddVolumeEntity struct {
	// EntityRuntimeID is the runtime ID of the volume. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the volume functions or looks.
	EntityMetadata map[string]any
	// EncodingIdentifier is the unique identifier for the volume. It must be of the form 'namespace:name', where
	// namespace cannot be 'minecraft'.
	EncodingIdentifier string
	// InstanceIdentifier is the identifier of a fog definition.
	InstanceIdentifier string
	// Bounds represent the volume's bounds. The first value is the minimum bounds, and the second value is the
	// maximum bounds.
	Bounds [2]protocol.BlockPos
	// Dimension is the dimension in which the volume exists.
	Dimension int32
	// EngineVersion is the engine version the entity is using, for example, '1.17.0'.
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
	w.UBlockPos(&pk.Bounds[0])
	w.UBlockPos(&pk.Bounds[1])
	w.Varint32(&pk.Dimension)
	w.String(&pk.EngineVersion)
}

// Unmarshal ...
func (pk *AddVolumeEntity) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.EntityRuntimeID)
	r.NBT(&pk.EntityMetadata, nbt.NetworkLittleEndian)
	r.String(&pk.EncodingIdentifier)
	r.String(&pk.InstanceIdentifier)
	r.UBlockPos(&pk.Bounds[0])
	r.UBlockPos(&pk.Bounds[1])
	r.Varint32(&pk.Dimension)
	r.String(&pk.EngineVersion)
}
