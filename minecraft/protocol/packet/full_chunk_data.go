package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// FullChunkData is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type FullChunkData struct {
	// ChunkX is the X coordinate of the chunk sent. (To translate a block's X to a chunk's X: x >> 4)
	ChunkX int32
	// ChunkZ is the Z coordinate of the chunk sent. (To translate a block's Z to a chunk's Z: z >> 4)
	ChunkZ int32
	// Data is a serialised string of chunk data. The chunk data is composed of multiple sub-chunks, each of
	// which carry a version which indicates the way they are serialised.
	Data string
}

// ID ...
func (*FullChunkData) ID() uint32 {
	return IDFullChunkData
}

// Marshal ...
func (pk *FullChunkData) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.ChunkX)
	_ = protocol.WriteVarint32(buf, pk.ChunkZ)
	_ = protocol.WriteString(buf, pk.Data)
}

// Unmarshal ...
func (pk *FullChunkData) Unmarshal(buf *bytes.Buffer) error {
	return ChainErr(
		protocol.Varint32(buf, &pk.ChunkX),
		protocol.Varint32(buf, &pk.ChunkZ),
		protocol.String(buf, &pk.Data),
	)
}
