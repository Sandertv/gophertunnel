package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubChunkRequest requests a specific sub-chunk from the server using the dimension and sub-chunk position.
type SubChunkRequest struct {
	// Dimension is the dimension of the sub-chunk.
	Dimension int32
	// Position is an absolute sub-chunk center point used as a base point for all sub-chunks requested. The X and Z
	// coordinates represent the chunk coordinates, while the Y coordinate is the absolute sub-chunk index.
	Position protocol.SubChunkPos
	// Offsets contains all requested offsets around the center point.
	Offsets [][3]byte
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
		w.Uint8(&offset[0])
		w.Uint8(&offset[1])
		w.Uint8(&offset[2])
	}
}

// Unmarshal ...
func (pk *SubChunkRequest) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Dimension)
	r.SubChunkPos(&pk.Position)

	var count uint32
	r.Uint32(&count)

	pk.Offsets = make([][3]byte, count)
	for i := uint32(0); i < count; i++ {
		var offset [3]byte
		r.Uint8(&offset[0])
		r.Uint8(&offset[1])
		r.Uint8(&offset[2])

		pk.Offsets[i] = offset
	}
}
