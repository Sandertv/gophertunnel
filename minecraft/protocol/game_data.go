package protocol

// BlockEntry is a block sent in the StartGame packet block runtime ID table. It holds a name and a metadata
// value of a block.
type BlockEntry struct {
	// Name is the name of the block. It looks like 'minecraft:stone'.
	Name string
	// RawPayload is the metadata value of the block. A lot of blocks only have 0 as data value, but some blocks
	// carry specific variants or properties encoded in the metadata.
	Data int16
	// LegacyID is the legacy, numerical ID of the block.
	LegacyID int16
}

// ItemEntry is an item sent in the StartGame item table. It holds a name and a legacy ID, which is used to
// point back to that name.
type ItemEntry struct {
	// Name if the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// LegacyID is the legacy ID of the item. It must point to either an existing item ID or a new one if it
	// seeks to implement a new item.
	LegacyID int16
}
