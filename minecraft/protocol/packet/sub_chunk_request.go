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
	Offsets [][3]int8
}

// ID ...
func (*SubChunkRequest) ID() uint32 {
	return IDSubChunkRequest
}

// Marshal ...
func (pk *SubChunkRequest) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Dimension)
	w.SubChunkPos(&pk.Position)

	count := uint32(len(pk.Offsets))
	w.Uint32(&count)
	for _, offset := range pk.Offsets {
		w.Int8(&offset[0])
		w.Int8(&offset[1])
		w.Int8(&offset[2])
	}
}

// Unmarshal ...
func (pk *SubChunkRequest) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Dimension)
	r.SubChunkPos(&pk.Position)

	var count uint32
	r.Uint32(&count)

	pk.Offsets = make([][3]int8, count)
	for i := uint32(0); i < count; i++ {
		var offset [3]int8
		r.Int8(&offset[0])
		r.Int8(&offset[1])
		r.Int8(&offset[2])

		pk.Offsets[i] = offset
	}
}
