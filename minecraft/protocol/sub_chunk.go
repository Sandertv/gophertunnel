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
	SubChunkResultUndefined = iota
	SubChunkResultSuccess
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

// Marshal encodes/decodes a SubChunkEntry.
func (x *SubChunkEntry) Marshal(r IO) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	hasRawPayload := x.RawPayload != nil
	r.Bool(&hasRawPayload)
	if hasRawPayload {
		r.ByteSlice(&x.RawPayload)
	} else {
		x.RawPayload = nil
	}
	r.Uint8(&x.HeightMapType)
	hasHeightMapData := x.HeightMapData != nil
	r.Bool(&hasHeightMapData)
	if hasHeightMapData {
		FuncSliceOfLen(r, 256, &x.HeightMapData, r.Int8)
	} else {
		x.HeightMapData = nil
	}
	r.Uint8(&x.RenderHeightMapType)
	hasRenderHeightMapData := x.RenderHeightMapData != nil
	r.Bool(&hasRenderHeightMapData)
	if hasRenderHeightMapData {
		FuncSliceOfLen(r, 256, &x.RenderHeightMapData, r.Int8)
	} else {
		x.RenderHeightMapData = nil
	}
	hasBlobHash := x.BlobHash != 0
	r.Bool(&hasBlobHash)
	if hasBlobHash {
		r.Uint64(&x.BlobHash)
	} else {
		x.BlobHash = 0
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
