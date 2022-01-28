package protocol

const (
	HeightMapDataNone byte = iota
	HeightMapDataHasData
	HeightMapDataTooHigh
	HeightMapDataTooLow
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
// requesting system introduced in v1.18.0.
type SubChunkEntry struct {
	// Offset contains the offset between the sub-chunk position and the center position.
	Offset [3]byte
	// Result is always one of the constants defined in the SubChunkResult constants.
	Result byte
	// RawPayload contains the serialized sub-chunk data.
	RawPayload []byte
	// HeightMapType is always one of the constants defined in the HeightMapData constants.
	HeightMapType byte
	// HeightMapData is the data for the height map.
	HeightMapData []byte
	// BlobHash is the hash of the blob.
	BlobHash uint64
}
