package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// RemoveVolumeEntity indicates a volume entity to be removed from server to client.
type RemoveVolumeEntity struct {
	// EntityRuntimeID ...
	EntityRuntimeID uint64
	// Dimension ...
	Dimension int32
}

// ID ...
func (*RemoveVolumeEntity) ID() uint32 {
	return IDRemoveVolumeEntity
}

// Marshal ...
func (pk *RemoveVolumeEntity) Marshal(w *protocol.Writer) {
	w.Uint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.Dimension)
}

// Unmarshal ...
func (pk *RemoveVolumeEntity) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.Dimension)
}
