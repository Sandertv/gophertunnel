package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubChunkRequest requests specific sub-chunks from the server using a center point.
type SubChunkRequest struct {
	// Dimension is the dimension of the sub-chunk.
	Dimension int32
	// Position is an absolute sub-chunk center point used as a base point for all sub-chunks requested. The X and Z
	// coordinates represent the chunk coordinates, while the Y coordinate is the absolute sub-chunk index.
	Position protocol.SubChunkPos
	// Offsets contains all requested offsets around the center point.
	Offsets []protocol.SubChunkOffset
}

// ID ...
func (*SubChunkRequest) ID() uint32 {
	return IDSubChunkRequest
}

// Marshal ...
func (pk *SubChunkRequest) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SubChunkRequest) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SubChunkRequest) marshal(r protocol.IO) {
	r.Varint32(&pk.Dimension)
	r.SubChunkPos(&pk.Position)
	protocol.SliceUint32Length(r, &pk.Offsets)
}
