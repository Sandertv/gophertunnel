package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LevelChunk is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type LevelChunk struct {
	// ChunkX is the X coordinate of the chunk sent. (To translate a block's X to a chunk's X: x >> 4)
	ChunkX int32
	// ChunkZ is the Z coordinate of the chunk sent. (To translate a block's Z to a chunk's Z: z >> 4)
	ChunkZ int32
	// SubChunkCount is the amount of sub chunks that are part of the chunk sent. Depending on if the cache
	// is enabled, a list of blob hashes will be sent, or, if disabled, the sub chunk data.
	SubChunkCount uint32
	// CacheEnabled specifies if the client blob cache should be enabled. This system is based on hashes of
	// blobs which are consistent and saved by the client in combination with that blob, so that the server
	// does not have the same chunk multiple times. If the client does not yet have a blob with the hash sent,
	// it will send a ClientCacheBlobStatus packet containing the hashes is does not have the data of.
	CacheEnabled bool
	// BlobHashes is a list of all blob hashes used in the chunk. It is composed of SubChunkCount + 1 hashes,
	// with the first SubChunkCount hashes being those of the sub chunks and the last one that of the biome
	// of the chunk.
	// If CacheEnabled is set to false, BlobHashes can be left empty.
	BlobHashes []int64
	// RawPayload is a serialised string of chunk data. The data held depends on if CacheEnabled is set to
	// true. If set to false, the payload is composed of multiple sub-chunks, each of which carry a version
	// which indicates the way they are serialised, followed by biomes, border blocks and tile entities. If
	// CacheEnabled is true, the payload consists out of the border blocks and tile entities only.
	RawPayload []byte
}

// ID ...
func (*LevelChunk) ID() uint32 {
	return IDLevelChunk
}

// Marshal ...
func (pk *LevelChunk) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.ChunkX)
	_ = protocol.WriteVarint32(buf, pk.ChunkZ)
	_ = protocol.WriteVaruint32(buf, pk.SubChunkCount)
	_ = binary.Write(buf, binary.LittleEndian, pk.CacheEnabled)
	if pk.CacheEnabled {
		_ = protocol.WriteVaruint32(buf, uint32(len(pk.BlobHashes)))
		for _, hash := range pk.BlobHashes {
			_ = binary.Write(buf, binary.LittleEndian, hash)
		}
	}
	_ = protocol.WriteByteSlice(buf, pk.RawPayload)
}

// Unmarshal ...
func (pk *LevelChunk) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Varint32(buf, &pk.ChunkX),
		protocol.Varint32(buf, &pk.ChunkZ),
		protocol.Varuint32(buf, &pk.SubChunkCount),
		binary.Read(buf, binary.LittleEndian, &pk.CacheEnabled),
	); err != nil {
		return err
	}
	if pk.CacheEnabled {
		var count uint32
		if err := protocol.Varuint32(buf, &count); err != nil {
			return err
		}
		pk.BlobHashes = make([]int64, count)
		for i := uint32(0); i < count; i++ {
			if err := binary.Read(buf, binary.LittleEndian, &pk.BlobHashes[i]); err != nil {
				return err
			}
		}
	}
	return protocol.ByteSlice(buf, &pk.RawPayload)
}
