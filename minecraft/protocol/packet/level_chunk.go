package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// LevelChunk is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically, a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type LevelChunk struct {
	// Position contains the X and Z coordinates of the chunk sent. You can convert a block coordinate to a chunk
	// coordinate by right-shifting it four bits.
	Position protocol.ChunkPos
	// SubChunkRequestMode specifies the sub-chunk request format. If it is not set, then sub-chunk requesting will not
	// be enabled. It is always one of the protocol.SubChunkRequestMode constants.
	SubChunkRequestMode byte
	// HighestSubChunk is the highest sub-chunk at the position that is not all air. It is only set if the
	// RequestMode is set to protocol.SubChunkRequestModeLimited.
	HighestSubChunk uint16
	// SubChunkCount is the amount of sub-chunks that are part of the chunk sent. Depending on if the cache
	// is enabled, a list of blob hashes will be sent, or, if disabled, the sub-chunk data.
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

// Marshal ...
func (pk *LevelChunk) Marshal(w *protocol.Writer) {
	w.ChunkPos(&pk.Position)
	switch pk.SubChunkRequestMode {
	case protocol.SubChunkRequestModeLegacy:
		w.Varuint32(&pk.SubChunkCount)
	case protocol.SubChunkRequestModeLimitless:
		limitlessFlag := uint32(math.MaxUint32)
		w.Varuint32(&limitlessFlag)
	case protocol.SubChunkRequestModeLimited:
		limitedFlag := uint32(math.MaxUint32 - 1)
		w.Varuint32(&limitedFlag)
		w.Uint16(&pk.HighestSubChunk)
	}

	w.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		l := uint32(len(pk.BlobHashes))
		w.Varuint32(&l)
		for _, hash := range pk.BlobHashes {
			w.Uint64(&hash)
		}
	}
	w.ByteSlice(&pk.RawPayload)
}

// Unmarshal ...
func (pk *LevelChunk) Unmarshal(r *protocol.Reader) {
	r.ChunkPos(&pk.Position)

	var potentialSubCount uint32
	r.Varuint32(&potentialSubCount)

	switch potentialSubCount {
	case math.MaxUint32:
		pk.SubChunkRequestMode = protocol.SubChunkRequestModeLimitless
	case math.MaxUint32 - 1:
		pk.SubChunkRequestMode = protocol.SubChunkRequestModeLimited
		r.Uint16(&pk.HighestSubChunk)
	default:
		pk.SubChunkCount = potentialSubCount
	}

	r.Bool(&pk.CacheEnabled)
	if pk.CacheEnabled {
		var count uint32
		r.Varuint32(&count)
		pk.BlobHashes = make([]uint64, count)
		for i := uint32(0); i < count; i++ {
			r.Uint64(&pk.BlobHashes[i])
		}
	}
	r.ByteSlice(&pk.RawPayload)
}
