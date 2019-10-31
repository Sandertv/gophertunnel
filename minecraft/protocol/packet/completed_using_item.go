package packet

import (
	"bytes"
	"encoding/binary"
)

const (
	UseItemEquipArmor = iota
	UseItemEat
	UseItemAttack
	UseItemConsume
	UseItemThrow
	UseItemShoot
	UseItemPlace
	UseItemFillBottle
	UseItemFillBucket
	UseItemPourBucket
	UseItemUseTool
	UseItemInteract
	UseItemRetrieved
	UseItemDyed
	UseItemTraded
)

// CompletedUsingItem is sent by the client when it finishes the using of an item, such as when it shoots an
// arrow or eats food. This packet holds a set of data that the server should not rely on, because it can
// obtain the information in more reliable ways.
type CompletedUsingItem struct {
	// UsedItemID is the item ID of the item that the client completed using. The server should not rely on
	// this, but should instead simply look at the held item.
	UsedItemID int16
	// UseMethod is the method of the using of the item that was completed. It is one of the constants that
	// may be found above.
	UseMethod int32
}

// ID ...
func (*CompletedUsingItem) ID() uint32 {
	return IDCompletedUsingItem
}

// Marshal ...
func (pk *CompletedUsingItem) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.UsedItemID)
	_ = binary.Write(buf, binary.LittleEndian, pk.UseMethod)
}

// Unmarshal ...
func (pk *CompletedUsingItem) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.UsedItemID),
		binary.Read(buf, binary.LittleEndian, &pk.UseMethod),
	)
}
