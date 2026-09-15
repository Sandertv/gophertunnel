package protocol

const (
	HeightMapDataNone = iota
	HeightMapDataHasData
	HeightMapDataTooHigh
	HeightMapDataTooLow
	HeightMapDataAllCopied
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
	// RawPayload contains the serialized sub-chunk data, if present.
	RawPayload Optional[[]byte]
	// HeightMapType is always one of the constants defined in the HeightMapData constants.
	HeightMapType byte
	// HeightMapData is the data for the height map, if present.
	HeightMapData Optional[[]int8]
	// RenderHeightMapType is always one of the constants defined in the HeightMapData constants.
	RenderHeightMapType byte
	// RenderHeightMapData is the data for the render height map, if present.
	RenderHeightMapData Optional[[]int8]
	// BlobHash is the hash of the blob, if present.
	BlobHash Optional[uint64]
}

// Marshal encodes/decodes a SubChunkEntry.
func (x *SubChunkEntry) Marshal(r IO) {
	Single(r, &x.Offset)
	r.Uint8(&x.Result)
	OptionalFunc(r, &x.RawPayload, r.ByteSlice)
	r.Uint8(&x.HeightMapType)
	OptionalFunc(r, &x.HeightMapData, func(data *[]int8) { heightMap(r, data) })
	r.Uint8(&x.RenderHeightMapType)
	OptionalFunc(r, &x.RenderHeightMapData, func(data *[]int8) { heightMap(r, data) })
	OptionalFunc(r, &x.BlobHash, r.Uint64)
}

// heightMap reads/writes a sub-chunk heightmap. It holds one height per
// column, but is not flat on the wire: each of the 16 rows carries its own
// length, so a map is 16 lengths and 256 heights. Callers supply the heights
// alone; the rows are framed here.
func heightMap(r IO, data *[]int8) {
	const rows, cols = 16, 16
	if _, reading := r.(sliceReader); reading {
		*data = make([]int8, rows*cols)
	} else if len(*data) != rows*cols {
		r.InvalidValue(len(*data), "heightmap", "must hold 256 heights")
		return
	}
	for row := 0; row < rows; row++ {
		n := uint32(cols)
		r.Varuint32(&n)
		if n != cols {
			r.InvalidValue(n, "heightmap row", "must hold 16 heights")
			return
		}
		for col := 0; col < cols; col++ {
			r.Int8(&(*data)[row*cols+col])
		}
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
