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
	w.Uint8(&pk.WindowID)
	w.Varint32(&pk.CraftingType)
	w.UUID(&pk.RecipeUUID)
	protocol.FuncSlice(w, &pk.Input, w.ItemInstance)
	protocol.FuncSlice(w, &pk.Output, w.ItemInstance)
}

// Unmarshal ...
func (pk *CraftingEvent) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.WindowID)
	r.Varint32(&pk.CraftingType)
	r.UUID(&pk.RecipeUUID)
	protocol.FuncSlice(r, &pk.Input, r.ItemInstance)
	protocol.FuncSlice(r, &pk.Output, r.ItemInstance)
}
