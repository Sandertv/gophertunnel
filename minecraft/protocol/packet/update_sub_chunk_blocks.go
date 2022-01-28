package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// UpdateSubChunkBlocks is essentially just UpdateBlock packet, however for a set of blocks in a sub-chunk.
type UpdateSubChunkBlocks struct {
	// Position is the position of the sub-chunk being referred to.
	Position protocol.SubChunkPos
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
	w.SubChunkPos(&pk.Position)

	blocksLen, extraLen := uint32(len(pk.Blocks)), uint32(len(pk.Extra))

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
	r.SubChunkPos(&pk.Position)

	var blocksLen uint32
	r.Varuint32(&blocksLen)

	pk.Blocks = make([]protocol.BlockChangeEntry, blocksLen)
	for i := uint32(0); i < blocksLen; i++ {
		protocol.BlockChange(r, &pk.Blocks[i])
	}

	var extraLen uint32
	r.Varuint32(&extraLen)

	pk.Extra = make([]protocol.BlockChangeEntry, extraLen)
	for i := uint32(0); i < extraLen; i++ {
		protocol.BlockChange(r, &pk.Extra[i])
	}
}
