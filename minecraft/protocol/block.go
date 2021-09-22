package protocol

import "github.com/sandertv/gophertunnel/minecraft/nbt"

// BlockEntry is an entry for a custom block found in the StartGame packet. The runtime ID of these custom
// block entries is based on the index they have in the block palette when the palette is ordered
// alphabetically.
type BlockEntry struct {
	// Name is the name of the custom block.
	Name string
	// Properties is a list of properties which, in combination with the name, specify a unique block.
	Properties map[string]interface{}
}

// Block reads a BlockEntry x from IO r.
func Block(r IO, x *BlockEntry) {
	r.String(&x.Name)
	r.NBT(&x.Properties, nbt.NetworkLittleEndian)
}

// BlockChangeEntry is used by the UpdateSubChunkBlocks packet.
type BlockChangeEntry struct {
	BlockPos
	// BlockRuntimeID is the runtime ID of the block.
	BlockRuntimeID uint32
	// Flags ...
	Flags uint32
	// SyncedUpdateEntityUniqueID ...
	SyncedUpdateEntityUniqueID uint64
	// SyncedUpdateType ...
	SyncedUpdateType uint32
}

// BlockChange reads/writes a BlockChangeEntry from an IO.
func BlockChange(r IO, x *BlockChangeEntry) {
	r.Varint32(&x.BlockPos[0])
	r.Varint32(&x.BlockPos[1])
	r.Varint32(&x.BlockPos[2])

	r.Varuint32(&x.BlockRuntimeID)
	r.Varuint32(&x.Flags)
	r.Varuint64(&x.SyncedUpdateEntityUniqueID)
	r.Varuint32(&x.SyncedUpdateType)
}
