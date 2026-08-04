package protocol

import "math"

const (
	HeightMapDataNone = iota
	HeightMapDataHasData
	HeightMapDataTooHigh
	HeightMapDataTooLow
	HeightMapDataAllCopied
)

const (
	SubChunkRequestModeLimitless = math.MaxUint32 - iota
	SubChunkRequestModeLimited
)

const (
	SubChunkResultSuccess = iota + 1
	SubChunkResultChunkNotFound
	SubChunkResultInvalidDimension
	SubChunkResultPlayerNotFound
	SubChunkResultIndexOutOfBounds
	SubChunkResultSuccessAllAir
)

// SubChunkEntry contains the data of a sub-chunk entry relative to a center sub chunk position, used for the sub-chunk
// requesting system introduced in v1.18.10.
type SubChunkEntry struct {
	// Offset contains the offset between the sub-chunk position and the center position.
	Offset SubChunkOffset
	// Result is always one of the constants defined in the SubChunkResult constants.
	Result byte
	// RawPayload contains the serialized sub-chunk data.
	RawPayload []byte
	// HeightMapType is always one of the constants defined in the HeightMapData constants.
	HeightMapType byte
	// HeightMapData is the data for the height map.
	HeightMapData []int8
	// RenderHeightMapType is always one of the constants defined in the HeightMapData constants.
	RenderHeightMapType byte
	// RenderHeightMapData is the data for the render height map.
	RenderHeightMapData []int8
	// BlobHash is the hash of the blob.
	BlobHash uint64
}

// Marshal encodes/decodes a SubChunkEntry assuming the blob cache is enabled.
func (x *SubChunkEntry) Marshal(r IO) {
	subChunkEntry(r, x, true)
}

// SubChunkEntryNoCache encodes/decodes a SubChunkEntry assuming the blob cache is not enabled.
func SubChunkEntryNoCache(r IO, x *SubChunkEntry) {
	subChunkEntry(r, x, false)
}

// subChunkEntry encodes/decodes a SubChunkEntry. Every field that used to be conditional now carries its own
// presence byte, so the presence read from the stream is what decides whether a field follows, while encoding
// derives it from the result and height map types it belongs to.
func subChunkEntry(r IO, x *SubChunkEntry, cache bool) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)

	payload := x.Result == SubChunkResultSuccess
	r.Bool(&payload)
	if payload {
		r.ByteSlice(&x.RawPayload)
	}
	r.Uint8(&x.HeightMapType)
	heights := x.HeightMapType == HeightMapDataHasData
	r.Bool(&heights)
	if heights {
		FuncSliceOfLen(r, 256, &x.HeightMapData, r.Int8)
	}
	r.Uint8(&x.RenderHeightMapType)
	renderHeights := x.RenderHeightMapType == HeightMapDataHasData
	r.Bool(&renderHeights)
	if renderHeights {
		FuncSliceOfLen(r, 256, &x.RenderHeightMapData, r.Int8)
	}
	r.Bool(&cache)
	if cache {
		r.Uint64(&x.BlobHash)
	}
}

// SubChunkOffset represents an offset from the base position of another sub chunk.
type SubChunkOffset [3]int8

// Marshal encodes/decodes a SubChunkOffset.
func (x *SubChunkOffset) Marshal(r IO) {
	r.Int8(&x[0])
	r.Int8(&x[1])
	r.Int8(&x[2])
}
