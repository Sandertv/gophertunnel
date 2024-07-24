package protocol

import "github.com/sandertv/gophertunnel/minecraft/nbt"

// BlockEntry is an entry for a custom block found in the StartGame packet. The runtime ID of these custom
// block entries is based on the index they have in the block palette when the palette is ordered
// alphabetically.
type BlockEntry struct {
	// Name is the name of the custom block.
	Name string
	// Properties is a list of properties which, in combination with the name, specify a unique block.
	Properties map[string]any
}

// Marshal encodes/decodes a BlockEntry.
func (x *BlockEntry) Marshal(r IO) {
	r.String(&x.Name)
	r.NBT(&x.Properties, nbt.NetworkLittleEndian)
}

// BlockChangeEntry is used by the UpdateSubChunkBlocks packet.
type BlockChangeEntry struct {
	BlockPos
	// BlockRuntimeID is the runtime ID of the block.
	BlockRuntimeID uint32
	// Flags is a combination of flags that specify the way the block is updated client-side.
	Flags uint32
	// SyncedUpdateEntityUniqueID  is the unique ID of the falling block entity that the block transitions to or that the entity transitions from if the block change entry is synced.
	SyncedUpdateEntityUniqueID uint64
	// SyncedUpdateType is the type of the transition that happened. It is either BlockToEntityTransition, when
	// a block placed becomes a falling entity, or EntityToBlockTransition, when a falling entity hits the
	// ground and becomes a solid block again.
	SyncedUpdateType uint32
}

// Marshal encodes/decodes a BlockChangeEntry.
func (x *BlockChangeEntry) Marshal(r IO) {
	r.UBlockPos(&x.BlockPos)
	r.Varuint32(&x.BlockRuntimeID)
	r.Varuint32(&x.Flags)
	r.Varuint64(&x.SyncedUpdateEntityUniqueID)
	r.Varuint32(&x.SyncedUpdateType)
}
