package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubChunkRequest requests specific sub-chunks from the server using a center point.
type SubChunkRequest struct {
	// Dimension is the dimension of the sub-chunk.
	Dimension int32
	// Offsets contains all requested offsets around the center point.
	Offsets []protocol.SubChunkOffset
	// Position is an absolute sub-chunk center point used as a base point for all sub-chunks requested. The X and Z
	// coordinates represent the chunk coordinates, while the Y coordinate is the absolute sub-chunk index.
	Position protocol.SubChunkPos
}

// ID ...
func (*SubChunkRequest) ID() uint32 {
	return IDSubChunkRequest
}

func (pk *SubChunkRequest) Marshal(io protocol.IO) {
	io.Varint32(&pk.Dimension)
	protocol.Slice(io, &pk.Offsets)
	io.Int32(&pk.Position[0])
	io.Int32(&pk.Position[1])
	io.Int32(&pk.Position[2])
}
