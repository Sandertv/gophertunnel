package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
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

// ItemInst reads an ItemInstance x from Buffer src.
func ItemInst(src *bytes.Buffer, x *ItemInstance) error {
	if err := chainErr(
		Varint32(src, &x.StackNetworkID),
		Item(src, &x.Stack),
	); err != nil {
		return err
	}
	if (x.Stack.Count == 0 || x.Stack.NetworkID == 0) && x.StackNetworkID != 0 {
		return fmt.Errorf("stack %#v is empty but network ID %v is non-zero", x.Stack, x.StackNetworkID)
	}
	return nil
}

// WriteItemInst writes an ItemInstance x to Buffer dst.
func WriteItemInst(dst *bytes.Buffer, x ItemInstance) error {
	if (x.Stack.Count == 0 || x.Stack.NetworkID == 0) && x.StackNetworkID != 0 {
		panic(fmt.Sprintf("stack %#v is empty but network ID %v is non-zero", x.Stack, x.StackNetworkID))
	}
	return chainErr(
		WriteVarint32(dst, x.StackNetworkID),
		WriteItem(dst, x.Stack),
	)
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

// Item reads an item stack from buffer src and stores it into item stack x.
func Item(src *bytes.Buffer, x *ItemStack) error {
	x.NBTData = make(map[string]interface{})
	if err := Varint32(src, &x.NetworkID); err != nil {
		return wrap(err)
	}
	if x.NetworkID == 0 {
		// The item was air, so there is no more data we should read for the item instance. After all, air
		// items aren't really anything.
		x.MetadataValue = 0
		x.Count = 0
		x.CanBePlacedOn, x.CanBreak = nil, nil
		return nil
	}
	var auxValue int32
	if err := Varint32(src, &auxValue); err != nil {
		return wrap(err)
	}
	x.MetadataValue = int16(auxValue >> 8)
	x.Count = int16(auxValue & 0xff)

	var userDataMarker int16
	if err := binary.Read(src, binary.LittleEndian, &userDataMarker); err != nil {
		return wrap(err)
	}
	if userDataMarker == -1 {
		var userDataVersion uint8
		if err := binary.Read(src, binary.LittleEndian, &userDataVersion); err != nil {
			return wrap(err)
		}
		switch userDataVersion {
		case 1:
			if err := nbt.NewDecoder(src).Decode(&x.NBTData); err != nil {
				return fmt.Errorf("%v: error decoding user data NBT: %v", callFrame(), err)
			}
		default:
			return fmt.Errorf("%v: unexpected item user data version %v", callFrame(), userDataVersion)
		}
	} else if userDataMarker != 0 {
		if userDataMarker < 0 {
			return fmt.Errorf("%v: invalid NBT length %v", callFrame(), userDataMarker)
		}
		nbtData := src.Next(int(userDataMarker))
		if err := nbt.UnmarshalEncoding(nbtData, &x.NBTData, nbt.LittleEndian); err != nil {
			return fmt.Errorf("%v: error decoding item NBT: %v", callFrame(), err)
		}
	}
	var count int32
	if err := Varint32(src, &count); err != nil {
		return wrap(err)
	}
	if count < 0 {
		return NegativeCountError{Type: "item can be placed on"}
	}
	if count > higherLimit {
		return LimitHitError{Limit: higherLimit, Type: "item can be placed on"}
	}
	x.CanBePlacedOn = make([]string, count)
	for i := int32(0); i < count; i++ {
		if err := String(src, &x.CanBePlacedOn[i]); err != nil {
			return wrap(err)
		}
	}
	if err := Varint32(src, &count); err != nil {
		return wrap(err)
	}
	if count < 0 {
		return NegativeCountError{Type: "item can break"}
	}
	if count > higherLimit {
		return LimitHitError{Limit: higherLimit, Type: "item can break"}
	}
	x.CanBreak = make([]string, count)
	for i := int32(0); i < count; i++ {
		if err := String(src, &x.CanBreak[i]); err != nil {
			return wrap(err)
		}
	}
	const shieldID = 513
	if x.NetworkID == shieldID {
		var blockingTick int64
		if err := Varint64(src, &blockingTick); err != nil {
			return wrap(err)
		}
	}
	return nil
}

// WriteItem writes an item stack x to buffer dst.
func WriteItem(dst *bytes.Buffer, x ItemStack) error {
	if err := WriteVarint32(dst, x.NetworkID); err != nil {
		return wrap(err)
	}
	if x.NetworkID == 0 {
		// The item was air, so there's no more data to follow. Return immediately.
		return nil
	}
	if err := WriteVarint32(dst, int32(x.MetadataValue<<8)|int32(x.Count)); err != nil {
		return wrap(err)
	}
	if len(x.NBTData) != 0 {
		// Write the item user data marker.
		if err := binary.Write(dst, binary.LittleEndian, int16(-1)); err != nil {
			return wrap(err)
		}
		// NBT version.
		if err := binary.Write(dst, binary.LittleEndian, byte(1)); err != nil {
			return wrap(err)
		}
		b, err := nbt.Marshal(x.NBTData)
		if err != nil {
			panic(fmt.Errorf("error writing item NBT of %#v: %w", x, err))
		}
		_, _ = dst.Write(b)
	} else {
		// If we write 0 for the marker, we don't have to write an empty compound tag.
		if err := binary.Write(dst, binary.LittleEndian, int16(0)); err != nil {
			return wrap(err)
		}
	}
	if err := WriteVarint32(dst, int32(len(x.CanBePlacedOn))); err != nil {
		return wrap(err)
	}
	for _, block := range x.CanBePlacedOn {
		if err := WriteString(dst, block); err != nil {
			return wrap(err)
		}
	}
	if err := WriteVarint32(dst, int32(len(x.CanBreak))); err != nil {
		return wrap(err)
	}
	for _, block := range x.CanBreak {
		if err := WriteString(dst, block); err != nil {
			return wrap(err)
		}
	}
	const shieldID = 513
	if x.NetworkID == shieldID {
		var blockingTick int64
		if err := WriteVarint64(dst, blockingTick); err != nil {
			return wrap(err)
		}
	}
	return nil
}

// RecipeIngredient reads an ItemStack x as a recipe ingredient from Buffer src.
func RecipeIngredient(src *bytes.Buffer, x *ItemStack) error {
	if err := Varint32(src, &x.NetworkID); err != nil {
		return nil
	}
	if x.NetworkID == 0 {
		return nil
	}
	var meta, count int32
	if err := Varint32(src, &meta); err != nil {
		return err
	}
	x.MetadataValue = int16(meta)
	if err := Varint32(src, &count); err != nil {
		return err
	}
	if count < 0 {
		return NegativeCountError{Type: "recipe ingredient count"}
	}
	x.Count = int16(count)
	return nil
}

// WriteRecipeIngredient writes an ItemStack x as a recipe ingredient to Buffer dst.
func WriteRecipeIngredient(dst *bytes.Buffer, x ItemStack) error {
	if err := WriteVarint32(dst, x.NetworkID); err != nil {
		return err
	}
	if x.NetworkID == 0 {
		return nil
	}
	return chainErr(
		WriteVarint32(dst, int32(x.MetadataValue)),
		WriteVarint32(dst, int32(x.Count)),
	)
}
