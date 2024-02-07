package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LevelChunk is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically, a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type LevelChunk struct {
	// Position contains the X and Z coordinates of the chunk sent. You can convert a block coordinate to a chunk
	// coordinate by right-shifting it four bits.
	Position protocol.ChunkPos
	// Dimension is the ID of the dimension that the chunk belongs to. This must always be set otherwise the
	// client will always assume the chunk is part of the overworld dimension.
	Dimension int32
	// HighestSubChunk is the highest sub-chunk at the position that is not all air. It is only set if the
	// SubChunkCount is set to protocol.SubChunkRequestModeLimited.
	HighestSubChunk uint16
	// SubChunkCount is the amount of sub-chunks that are part of the chunk
	// sent. Depending on if the cache is enabled, a list of blob hashes will be
	// sent, or, if disabled, the sub-chunk data. SubChunkCount may be set to
	// protocol.SubChunkRequestModeLimited or
	// protocol.SubChunkRequestModeLimitless to prompt the client to send a
	// SubChunkRequest in response. If this field is set to
	// protocol.SubChunkRequestModeLimited, HighestSubChunk is used.
	SubChunkCount uint32
	// CacheEnabled specifies if the client blob cache should be enabled. This system is based on hashes of
	// blobs which are consistent and saved by the client in combination with that blob, so that the server
	// does not have the same chunk multiple times. If the client does not yet have a blob with the hash sent,
	// it will send a ClientCacheBlobStatus packet containing the hashes is does not have the data of.
	CacheEnabled bool
	// BlobHashes is a list of all blob hashes used in the chunk. It is composed of SubChunkCount + 1 hashes,
	// with the first SubChunkCount hashes being those of the sub-chunks and the last one that of the biome
	// of the chunk.
	// If CacheEnabled is set to false, BlobHashes can be left empty.
	BlobHashes []uint64
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

func (pk *LevelChunk) Marshal(io protocol.IO) {
	io.ChunkPos(&pk.Position)
	io.Varint32(&pk.Dimension)
	io.Varuint32(&pk.SubChunkCount)
	if pk.SubChunkCount == protocol.SubChunkRequestModeLimited {
		io.Uint16(&pk.HighestSubChunk)
	}
	io.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		protocol.FuncSlice(io, &pk.BlobHashes, io.Uint64)
	}
	io.ByteSlice(&pk.RawPayload)
}
