package protocol

// ChunkPos is the position of a chunk. It is composed of two integers, and is typically written as either
// two varint32s.
type ChunkPos [2]int32

// X returns the X coordinate of the chunk position. It is equivalent to BlockPos[0].
func (pos ChunkPos) X() int32 {
	return pos[0]
}

// Z returns the Z coordinate of the chunk position. It is equivalent to BlockPos[2].
func (pos ChunkPos) Z() int32 {
	return pos[1]
}

// SubChunkPos is the position of a subchunk. The X and Z coordinates are the coordinates of the chunk, and the Y
// coordinate is the absolute subchunk index.
type SubChunkPos [3]int32

// X returns the X coordinate of the subchunk position. It is equivalent to SubChunkPos[0].
func (pos SubChunkPos) X() int32 {
	return pos[0]
}

// Y returns the Y coordinate of the subchunk position. It is equivalent to SubChunkPos[1].
func (pos SubChunkPos) Y() int32 {
	return pos[1]
}

// Z returns the Z coordinate of the subchunk position. It is equivalent to SubChunkPos[2].
func (pos SubChunkPos) Z() int32 {
	return pos[2]
}
