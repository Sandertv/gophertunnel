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

func (pk *SubChunk) Marshal(io protocol.IO) {
	io.Bool(&pk.CacheEnabled)
	io.Varint32(&pk.Dimension)
	io.SubChunkPos(&pk.Position)
	if pk.CacheEnabled {
		protocol.SliceUint32Length(io, &pk.SubChunkEntries)
	} else {
		protocol.FuncIOSliceUint32Length(io, &pk.SubChunkEntries, protocol.SubChunkEntryNoCache)
	}
}
