package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveVolumeEntity indicates a volume entity to be removed from server to client.
type RemoveVolumeEntity struct {
	// EntityRuntimeID ...
	EntityRuntimeID uint32
	// Dimension ...
	Dimension int32
}

// ID ...
func (*RemoveVolumeEntity) ID() uint32 {
	return IDRemoveVolumeEntity
}

func (pk *RemoveVolumeEntity) Marshal(io protocol.IO) {
	protocol.ActorRuntimeIDUint32(io, &pk.EntityRuntimeID)
	io.Varint32(&pk.Dimension)
}
