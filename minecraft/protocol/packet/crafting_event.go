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
	CraftingType int32 // TODO: Figure out the options for this.
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
func (pk *CraftingEvent) Unmarshal(buf *bytes.Buffer) error {
	var length uint32
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.WindowID),
		protocol.Varint32(buf, &pk.CraftingType),
		protocol.UUID(buf, &pk.RecipeUUID),
		protocol.Varuint32(buf, &length),
	); err != nil {
		return err
	}
	pk.Input = make([]protocol.ItemStack, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.Item(buf, &pk.Input[i]); err != nil {
			return err
		}
	}
	if err := protocol.Varuint32(buf, &length); err != nil {
		return err
	}
	pk.Output = make([]protocol.ItemStack, length)
	for i := uint32(0); i < length; i++ {
		if err := protocol.Item(buf, &pk.Input[i]); err != nil {
			return err
		}
	}
	return nil
}
