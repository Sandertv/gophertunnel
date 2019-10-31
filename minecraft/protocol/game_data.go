package protocol

// ItemEntry is an item sent in the StartGame item table. It holds a name and a legacy ID, which is used to
// point back to that name.
type ItemEntry struct {
	// Name if the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// LegacyID is the legacy ID of the item. It must point to either an existing item ID or a new one if it
	// seeks to implement a new item.
	LegacyID int16
}
