package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AddVolumeEntity sends a volume entity's definition and components from server to client.
type AddVolumeEntity struct {
	// EntityRuntimeID ...
	EntityRuntimeID uint64
	// VolumeEntityData ...
	VolumeEntityData map[string]interface{}
}

// ID ...
func (*AddVolumeEntity) ID() uint32 {
	return IDAddVolumeEntity
}

// Marshal ...
func (pk *AddVolumeEntity) Marshal(w *protocol.Writer) {
	w.Uint64(&pk.EntityRuntimeID)
	w.NBT(&pk.VolumeEntityData, nbt.NetworkLittleEndian)
}

// Unmarshal ...
func (pk *AddVolumeEntity) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.EntityRuntimeID)
	r.NBT(&pk.VolumeEntityData, nbt.NetworkLittleEndian)
}
