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
	w.Bool(&pk.CacheEnabled)
	w.Varint32(&pk.Dimension)
	w.SubChunkPos(&pk.Position)

	count := uint32(len(pk.SubChunkEntries))
	w.Uint32(&count)
	for _, entry := range pk.SubChunkEntries {
		w.Int8(&entry.Offset[0])
		w.Int8(&entry.Offset[1])
		w.Int8(&entry.Offset[2])

		w.Uint8(&entry.Result)
		if !pk.CacheEnabled || entry.Result != protocol.SubChunkResultSuccessAllAir {
			w.ByteSlice(&entry.RawPayload)
		}

		w.Uint8(&entry.HeightMapType)
		if entry.HeightMapType == protocol.HeightMapDataHasData {
			for i := 0; i < 256; i++ {
				w.Int8(&entry.HeightMapData[i])
			}
		}

		if pk.CacheEnabled {
			w.Uint64(&entry.BlobHash)
		}
	}
}

// Unmarshal ...
func (pk *SubChunk) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.CacheEnabled)
	r.Varint32(&pk.Dimension)
	r.SubChunkPos(&pk.Position)

	var count uint32
	r.Uint32(&count)

	pk.SubChunkEntries = make([]protocol.SubChunkEntry, count)
	for i := uint32(0); i < count; i++ {
		var entry protocol.SubChunkEntry

		r.Int8(&entry.Offset[0])
		r.Int8(&entry.Offset[1])
		r.Int8(&entry.Offset[2])

		r.Uint8(&entry.Result)
		if !pk.CacheEnabled || entry.Result != protocol.SubChunkResultSuccessAllAir {
			r.ByteSlice(&entry.RawPayload)
		}

		r.Uint8(&entry.HeightMapType)
		if entry.HeightMapType == protocol.HeightMapDataHasData {
			entry.HeightMapData = make([]int8, 256)
			for i := 0; i < 256; i++ {
				r.Int8(&entry.HeightMapData[i])
			}
		}

		if pk.CacheEnabled {
			r.Uint64(&entry.BlobHash)
		}

		pk.SubChunkEntries[i] = entry
	}
}
