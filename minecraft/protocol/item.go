package protocol

import "github.com/sandertv/gophertunnel/minecraft/nbt"

const (
	ItemEntryVersionLegacy = iota
	ItemEntryVersionDataDriven
	ItemEntryVersionNone
)

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

// ItemStack represents an item instance/stack over network. It has a network ID and a metadata value that
// define its type.
type ItemStack struct {
	ItemType
	// BlockRuntimeID ...
	BlockRuntimeID int32
	// Count is the count of items that the item stack holds.
	Count uint16
	// NBTData is a map that is serialised to its NBT representation when sent in a packet.
	NBTData map[string]any
	// CanBePlacedOn is a list of block identifiers like 'minecraft:stone' which the item, if it is an item
	// that can be placed, can be placed on top of.
	CanBePlacedOn []string
	// CanBreak is a list of block identifiers like 'minecraft:dirt' that the item is able to break.
	CanBreak []string
	// HasNetworkID ...
	HasNetworkID bool
}

// ItemType represents a consistent combination of network ID and metadata value of an item. It cannot usually
// be changed unless a new item is obtained.
type ItemType struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int32
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue uint32
}

// ItemEntry is an item sent in the StartGame item table. It holds a name and a legacy ID, which is used to
// point back to that name.
type ItemEntry struct {
	// Name if the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// RuntimeID is the ID that is used to identify the item over network. After sending all items in the
	// StartGame packet, items will then be identified using these numerical IDs.
	RuntimeID int16
	// ComponentBased specifies if the item was created using components, meaning the item is a custom item.
	ComponentBased bool
	// Version is the version of the item entry which is used by the client to determine how to handle the
	// item entry. It is one of the constants above.
	Version int32
	// Data is a map containing the components and properties of the item, if the item is component based.
	Data map[string]any
}

// Marshal encodes/decodes an ItemEntry.
func (x *ItemEntry) Marshal(r IO) {
	r.String(&x.Name)
	r.Int16(&x.RuntimeID)
	r.Bool(&x.ComponentBased)
	r.Varint32(&x.Version)
	r.NBT(&x.Data, nbt.NetworkLittleEndian)
}

// MaterialReducerOutput is an output from a material reducer.
type MaterialReducerOutput struct {
	// NetworkID is the network ID of the output.
	NetworkID int32
	// Count is the quantity of the output.
	Count int32
}

// Marshal encodes/decodes a MaterialReducerOutput.
func (x *MaterialReducerOutput) Marshal(r IO) {
	r.Varint32(&x.NetworkID)
	r.Varint32(&x.Count)
}

// MaterialReducer is a craft in a material reducer block in education edition.
type MaterialReducer struct {
	// InputItem is the starting item.
	InputItem ItemType
	// Outputs contain all outputting items.
	Outputs []MaterialReducerOutput
}
