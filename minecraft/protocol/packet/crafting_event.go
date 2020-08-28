package packet

import (
	"bytes"
	"encoding/binary"
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
	Input []protocol.ItemStack
	// Output is a list of items that were obtained as a result of crafting the recipe.
	Output []protocol.ItemStack
}

// ID ...
func (*CraftingEvent) ID() uint32 {
	return IDCraftingEvent
}

// Marshal ...
func (pk *CraftingEvent) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.WindowID)
	_ = protocol.WriteVarint32(buf, pk.CraftingType)
	_ = protocol.WriteUUID(buf, pk.RecipeUUID)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Input)))
	for _, input := range pk.Input {
		_ = protocol.WriteItem(buf, input)
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Output)))
	for _, output := range pk.Output {
		_ = protocol.WriteItem(buf, output)
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

	pk.Input = make([]protocol.ItemStack, length)
	for i := uint32(0); i < length; i++ {
		protocol.Item(r, &pk.Input[i])
	}
	r.Varuint32(&length)
	r.LimitUint32(length, 64)

	pk.Output = make([]protocol.ItemStack, length)
	for i := uint32(0); i < length; i++ {
		protocol.Item(r, &pk.Output[i])
	}
}
