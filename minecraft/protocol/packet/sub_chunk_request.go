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

func (pk *SubChunkRequest) Marshal(io protocol.IO) {
	io.Varint32(&pk.Dimension)
	io.SubChunkPos(&pk.Position)
	protocol.SliceUint32Length(io, &pk.Offsets)
}
