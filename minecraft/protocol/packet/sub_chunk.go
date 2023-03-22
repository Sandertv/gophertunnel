package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SubChunk sends data about multiple sub-chunks around a center point.
type SubChunk struct {
	// CacheEnabled is whether the sub-chunk caching is enabled or not.
	CacheEnabled bool
	// Dimension is the dimension the sub-chunks are in.
	Dimension int32
	// Position is an absolute sub-chunk center point that every SubChunkRequest uses as a reference.
	Position protocol.SubChunkPos
	// SubChunkEntries contains sub-chunk entries relative to the center point.
	SubChunkEntries []protocol.SubChunkEntry
}

// ID ...
func (*SubChunk) ID() uint32 {
	return IDSubChunk
}

// Marshal ...
func (pk *SubChunk) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SubChunk) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SubChunk) marshal(r protocol.IO) {
	r.Bool(&pk.CacheEnabled)
	r.Varint32(&pk.Dimension)
	r.SubChunkPos(&pk.Position)
	if pk.CacheEnabled {
		protocol.SliceUint32Length(r, &pk.SubChunkEntries)
	} else {
		protocol.FuncIOSliceUint32Length(r, &pk.SubChunkEntries, protocol.SubChunkEntryNoCache)
	}
}
