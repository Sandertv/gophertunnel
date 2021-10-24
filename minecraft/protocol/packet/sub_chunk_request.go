package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubChunkRequest requests a specific sub chunk from the server using the dimension and sub chunk position.
type SubChunkRequest struct {
	// Dimension is the dimension of the sub chunk.
	Dimension int32
	// SubChunkX, SubChunkY, and SubChunkZ help identify the sub chunk.
	SubChunkX, SubChunkY, SubChunkZ int32
}

// ID ...
func (*SubChunkRequest) ID() uint32 {
	return IDSubChunkRequest
}

// Marshal ...
func (pk *SubChunkRequest) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Dimension)
	w.Varint32(&pk.SubChunkX)
	w.Varint32(&pk.SubChunkY)
	w.Varint32(&pk.SubChunkZ)
}

// Unmarshal ...
func (pk *SubChunkRequest) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Dimension)
	r.Varint32(&pk.SubChunkX)
	r.Varint32(&pk.SubChunkY)
	r.Varint32(&pk.SubChunkZ)
}
