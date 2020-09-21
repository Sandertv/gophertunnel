package protocol

// ItemInstance represents a unique instance of an item stack. These instances carry a specific network ID
// that is persistent for the stack.
type ItemInstance struct {
	// StackNetworkID is the network ID of the item stack. If the stack is empty, 0 is always written for this
	// field. If not, the field should be set to 1 if the server authoritative inventories are disabled in the
	// StartGame packet, or to a unique stack ID if it is enabled.
	StackNetworkID int32
	// Stack is the actual item stack of the item instance.
	Stack ItemStack
}

// ItemInst reads/writes an ItemInstance x using IO r.
func ItemInst(r IO, x *ItemInstance) {
	r.Varint32(&x.StackNetworkID)
	r.Item(&x.Stack)
	if (x.Stack.Count == 0 || x.Stack.NetworkID == 0) && x.StackNetworkID != 0 {
		r.InvalidValue(x.StackNetworkID, "stack network ID", "stack is empty but network ID is non-zero")
	}
}

// ItemStack represents an item instance/stack over network. It has a network ID and a metadata value that
// define its type.
type ItemStack struct {
	ItemType
	// Count is the count of items that the item stack holds.
	Count int16
	// NBTData is a map that is serialised to its NBT representation when sent in a packet.
	NBTData map[string]interface{}
	// CanBePlacedOn is a list of block identifiers like 'minecraft:stone' which the item, if it is an item
	// that can be placed, can be placed on top of.
	CanBePlacedOn []string
	// CanBreak is a list of block identifiers like 'minecraft:dirt' that the item is able to break.
	CanBreak []string
}

// ItemType represents a consistent combination of network ID and metadata value of an item. It cannot usually
// be changed unless a new item is obtained.
type ItemType struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int32
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int16
}

// RecipeIngredientItem represents an item that may be used as a recipe ingredient.
type RecipeIngredientItem struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int32
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int32
	// Count is the count of items that the recipe ingredient is required to have.
	Count int32
}

// RecipeIngredient reads/writes a RecipeIngredientItem x using IO r.
func RecipeIngredient(r IO, x *RecipeIngredientItem) {
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		return
	}
	r.Varint32(&x.MetadataValue)
	r.Varint32(&x.Count)
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
