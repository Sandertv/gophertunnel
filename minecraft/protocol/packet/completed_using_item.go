package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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

// CompletedUsingItem is sent by the server to tell the client that it should be done using the item it is
// currently using.
type CompletedUsingItem struct {
	// UsedItemID is the item ID of the item that the client completed using. This should typically be the
	// ID of the item held in the hand.
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
func (pk *CompletedUsingItem) Marshal(w *protocol.Writer) {
	w.Int16(&pk.UsedItemID)
	w.Int32(&pk.UseMethod)
}

// Unmarshal ...
func (pk *CompletedUsingItem) Unmarshal(r *protocol.Reader) {
	r.Int16(&pk.UsedItemID)
	r.Int32(&pk.UseMethod)
}
