package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// UpdateSubChunkBlocks is essentially just UpdateBlock packet, however for a set of blocks in a sub-chunk.
type UpdateSubChunkBlocks struct {
	// Position is the block position of the sub-chunk being referred to.
	// If Position is (x, y, z), then (x>>4, y>>4, z>>4) is the corresponding sub chunk position.
	Position protocol.BlockPos
	// Blocks contains each updated block change entry.
	Blocks []protocol.BlockChangeEntry
	// Extra contains each updated block change entry for the second layer, usually for waterlogged blocks.
	Extra []protocol.BlockChangeEntry
}

// ID ...
func (*UpdateSubChunkBlocks) ID() uint32 {
	return IDUpdateSubChunkBlocks
}

func (pk *UpdateSubChunkBlocks) Marshal(io protocol.IO) {
	io.UBlockPos(&pk.Position)
	protocol.Slice(io, &pk.Blocks)
	protocol.Slice(io, &pk.Extra)
}
