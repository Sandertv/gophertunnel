package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	UseItemUnknown           = -1
	UseItemEquipArmor        = 0
	UseItemEat               = 1
	UseItemAttack            = 2
	UseItemConsume           = 3
	UseItemThrow             = 4
	UseItemShoot             = 5
	UseItemPlace             = 6
	UseItemFillBottle        = 7
	UseItemFillBucket        = 8
	UseItemPourBucket        = 9
	UseItemUseTool           = 10
	UseItemInteract          = 11
	UseItemRetrieved         = 12
	UseItemDyed              = 13
	UseItemTraded            = 14
	UseItemBrushingCompleted = 15
	UseItemOpenedVault       = 16
	UseItemCount             = 17
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

func (pk *CompletedUsingItem) Marshal(io protocol.IO) {
	io.Int16(&pk.UsedItemID)
	io.Int32(&pk.UseMethod)
}
