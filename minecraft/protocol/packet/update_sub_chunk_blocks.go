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
	protocol.Slice(w, &pk.Blocks)
	protocol.Slice(w, &pk.Extra)
}

// Unmarshal ...
func (pk *UpdateSubChunkBlocks) Unmarshal(r *protocol.Reader) {
	r.SubChunkPos(&pk.Position)
	protocol.Slice(r, &pk.Blocks)
	protocol.Slice(r, &pk.Extra)
}
