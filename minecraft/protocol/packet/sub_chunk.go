package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SubChunkRequestResultUndefined int32 = iota
	SubChunkRequestResultSuccess
	SubChunkRequestResultChunkNotFound
	SubChunkRequestResultInvalidDimension
	SubChunkRequestResultPlayerNotFound
	SubChunkRequestResultIndexOutOfBounds
)

const (
	HeightMapDataTypeNone byte = iota
	HeightMapDataTypeHasData
	HeightMapDataTypeTooHigh
	HeightMapDataTypeTooLow
)

// SubChunk sends sub chunk data about a specific chunk using its position and dimension.
type SubChunk struct {
	// Dimension is the dimension of the sub chunk.
	Dimension int32
	// SubChunkX, SubChunkY, and SubChunkZ help identify the sub chunk.
	SubChunkX, SubChunkY, SubChunkZ int32
	// Data is the actual sub chunk data, such as the blocks.
	Data []byte
	// RequestResult is one of the values above.
	RequestResult int32
	// HeightMapType is one of the values above.
	HeightMapType byte
	// HeightMapData is the data for the height map.
	HeightMapData []byte
	// CacheEnabled is whether the sub chunk caching is enabled or not.
	CacheEnabled bool
	// BlobHash is the hash of the blob.
	BlobHash uint64
}

// ID ...
func (*SubChunk) ID() uint32 {
	return IDSubChunk
}

// Marshal ...
func (pk *SubChunk) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Dimension)
	w.Varint32(&pk.SubChunkX)
	w.Varint32(&pk.SubChunkY)
	w.Varint32(&pk.SubChunkZ)
	w.ByteSlice(&pk.Data)
	w.Varint32(&pk.RequestResult)

	w.Uint8(&pk.HeightMapType)
	if pk.HeightMapType == HeightMapDataTypeHasData {
		for i := 0; i < 256; i++ {
			w.Uint8(&pk.HeightMapData[i])
		}
	}

	w.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		w.Uint64(&pk.BlobHash)
	}
}

// Unmarshal ...
func (pk *SubChunk) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Dimension)
	r.Varint32(&pk.SubChunkX)
	r.Varint32(&pk.SubChunkY)
	r.Varint32(&pk.SubChunkZ)
	r.ByteSlice(&pk.Data)
	r.Varint32(&pk.RequestResult)

	r.Uint8(&pk.HeightMapType)
	if pk.HeightMapType == HeightMapDataTypeHasData {
		pk.HeightMapData = make([]uint8, 256)
		for i := 0; i < 256; i++ {
			r.Uint8(&pk.HeightMapData[i])
		}
	}

	r.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		r.Uint64(&pk.BlobHash)
	}
}
