package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// UpdateSubChunkBlocks is essentially just UpdateBlock packet, however for a set of blocks in a sub chunk.
type UpdateSubChunkBlocks struct {
	// ChunkX, ChunkY, and ChunkZ help identify the sub chunk.
	ChunkX, ChunkY, ChunkZ int32
	// Blocks contains each updated block change entry.
	Blocks []protocol.BlockChangeEntry
	// Extra contains each updated block change entry for the second layer, usually for waterlogged blocks.
	Extra []protocol.BlockChangeEntry
}

// ID ...
func (*UpdateSubChunkBlocks) ID() uint32 {
	return IDUpdateSubChunkBlocks
}

// Marshal ...
func (pk *UpdateSubChunkBlocks) Marshal(w *protocol.Writer) {
	blocksLen, extraLen := uint32(len(pk.Blocks)), uint32(len(pk.Extra))

	w.Varint32(&pk.ChunkX)
	w.Varint32(&pk.ChunkY)
	w.Varint32(&pk.ChunkZ)

	w.Varuint32(&blocksLen)
	for _, entry := range pk.Blocks {
		protocol.BlockChange(w, &entry)
	}

	w.Varuint32(&extraLen)
	for _, entry := range pk.Extra {
		protocol.BlockChange(w, &entry)
	}
}

// Unmarshal ...
func (pk *UpdateSubChunkBlocks) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.ChunkX)
	r.Varint32(&pk.ChunkY)
	r.Varint32(&pk.ChunkZ)

	var blocksLen, extraLen uint32

	r.Varuint32(&blocksLen)

	pk.Blocks = make([]protocol.BlockChangeEntry, blocksLen)
	for i := uint32(0); i < blocksLen; i++ {
		protocol.BlockChange(r, &pk.Blocks[i])
	}

	r.Varuint32(&extraLen)

	pk.Extra = make([]protocol.BlockChangeEntry, extraLen)
	for i := uint32(0); i < extraLen; i++ {
		protocol.BlockChange(r, &pk.Extra[i])
	}
}
