package packet

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CraftingEvent is sent by the client when it crafts a particular item. Note that this packet may be fully
// ignored, as the InventoryTransaction packet provides all the information required.
type CraftingEvent struct {
	// WindowID is the ID representing the window that the player crafted in.
	WindowID byte
	// CraftingType is a type that indicates the way the crafting was done, for example if a crafting table
	// was used.
	// TODO: Find out the options of the CraftingType field in the CraftingEvent packet.
	CraftingType int32
	// RecipeUUID is the UUID of the recipe that was crafted. It points to the UUID of the recipe that was
	// sent earlier in the CraftingData packet.
	RecipeUUID uuid.UUID
	// Input is a list of items that the player put into the recipe so that it could create the Output items.
	// These items are consumed in the process.
	Input []protocol.ItemInstance
	// Output is a list of items that were obtained as a result of crafting the recipe.
	Output []protocol.ItemInstance
}

// ID ...
func (*CraftingEvent) ID() uint32 {
	return IDCraftingEvent
}

// Marshal ...
func (pk *CraftingEvent) Marshal(w *protocol.Writer) {
	inputLen, outputLen := uint32(len(pk.Input)), uint32(len(pk.Output))
	w.Uint8(&pk.WindowID)
	w.Varint32(&pk.CraftingType)
	w.UUID(&pk.RecipeUUID)
	w.Varuint32(&inputLen)
	for _, input := range pk.Input {
		w.ItemInstance(&input)
	}
	w.Varuint32(&outputLen)
	for _, output := range pk.Output {
		w.ItemInstance(&output)
	}
}

// Unmarshal ...
func (pk *CraftingEvent) Unmarshal(r *protocol.Reader) {
	var length uint32
	r.Uint8(&pk.WindowID)
	r.Varint32(&pk.CraftingType)
	r.UUID(&pk.RecipeUUID)
	r.Varuint32(&length)
	r.LimitUint32(length, 64)

	pk.Input = make([]protocol.ItemInstance, length)
	for i := uint32(0); i < length; i++ {
		r.ItemInstance(&pk.Input[i])
	}
	r.Varuint32(&length)
	r.LimitUint32(length, 64)

	pk.Output = make([]protocol.ItemInstance, length)
	for i := uint32(0); i < length; i++ {
		r.ItemInstance(&pk.Output[i])
	}
}
