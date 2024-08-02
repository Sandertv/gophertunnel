package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

const (
	InventoryActionSourceContainer = 0
	InventoryActionSourceWorld     = 2
	InventoryActionSourceCreative  = 3
	InventoryActionSourceTODO      = 99999
)

const (
	WindowIDInventory = 0
	WindowIDOffHand   = 119
	WindowIDArmour    = 120
	WindowIDUI        = 124
)

// InventoryAction represents a single action that took place during an inventory transaction. On itself, this
// inventory action is always unbalanced: It must be combined with other actions in an inventory transaction
// to form a balanced transaction.
type InventoryAction struct {
	// SourceType is the source type of the inventory action. It is one of the constants above.
	SourceType uint32
	// WindowID is the ID of the window that the client has opened. The window ID is not set if the SourceType
	// is InventoryActionSourceWorld.
	WindowID int32
	// SourceFlags is a combination of flags that is only set if the SourceType is InventoryActionSourceWorld.
	SourceFlags uint32
	// InventorySlot is the slot in which the action took place. Each action only describes the change of item
	// in a single slot.
	InventorySlot uint32
	// OldItem is the item that was present in the slot before the inventory action. It should be checked by
	// the server to ensure the inventories were not out of sync.
	OldItem ItemInstance
	// NewItem is the new item that was put in the InventorySlot that the OldItem was in. It must be checked
	// in combination with other inventory actions to ensure that the transaction is balanced.
	NewItem ItemInstance
}

// Marshal encodes/decodes an InventoryAction.
func (x *InventoryAction) Marshal(r IO) {
	r.Varuint32(&x.SourceType)
	switch x.SourceType {
	case InventoryActionSourceContainer, InventoryActionSourceTODO:
		r.Varint32(&x.WindowID)
	case InventoryActionSourceWorld:
		r.Varuint32(&x.SourceFlags)
	}
	r.Varuint32(&x.InventorySlot)
	r.ItemInstance(&x.OldItem)
	r.ItemInstance(&x.NewItem)
}

const (
	InventoryTransactionTypeNormal = iota
	InventoryTransactionTypeMismatch
	InventoryTransactionTypeUseItem
	InventoryTransactionTypeUseItemOnEntity
	InventoryTransactionTypeReleaseItem
)

// InventoryTransactionData represents an object that holds data specific to an inventory transaction type.
// The data it holds depends on the type.
type InventoryTransactionData interface {
	// Marshal encodes/decodes a serialised inventory transaction data object.
	Marshal(r IO)
}

// lookupTransactionData looks up inventory transaction data for the ID passed.
func lookupTransactionData(id uint32, x *InventoryTransactionData) bool {
	switch id {
	case InventoryTransactionTypeNormal:
		*x = &NormalTransactionData{}
	case InventoryTransactionTypeMismatch:
		*x = &MismatchTransactionData{}
	case InventoryTransactionTypeUseItem:
		*x = &UseItemTransactionData{}
	case InventoryTransactionTypeUseItemOnEntity:
		*x = &UseItemOnEntityTransactionData{}
	case InventoryTransactionTypeReleaseItem:
		*x = &ReleaseItemTransactionData{}
	default:
		return false
	}
	return true
}

// lookupTransactionDataType looks up an ID for a specific transaction data.
func lookupTransactionDataType(x InventoryTransactionData, id *uint32) bool {
	switch x.(type) {
	case *NormalTransactionData:
		*id = InventoryTransactionTypeNormal
	case *MismatchTransactionData:
		*id = InventoryTransactionTypeMismatch
	case *UseItemTransactionData:
		*id = InventoryTransactionTypeUseItem
	case *UseItemOnEntityTransactionData:
		*id = InventoryTransactionTypeUseItemOnEntity
	case *ReleaseItemTransactionData:
		*id = InventoryTransactionTypeReleaseItem
	default:
		return false
	}
	return true
}

// NormalTransactionData represents an inventory transaction data object for normal transactions, such as
// crafting. It has no content.
type NormalTransactionData struct{}

// MismatchTransactionData represents a mismatched inventory transaction's data object.
type MismatchTransactionData struct{}

const (
	UseItemActionClickBlock = iota
	UseItemActionClickAir
	UseItemActionBreakBlock
)

const (
	TriggerTypeUnknown = iota
	TriggerTypePlayerInput
	TriggerTypeSimulationTick
)

const (
	ClientPredictionFailure = iota
	ClientPredictionSuccess
)

// UseItemTransactionData represents an inventory transaction data object sent when the client uses an item on
// a block.
type UseItemTransactionData struct {
	// LegacyRequestID is an ID that is only non-zero at times when sent by the client. The server should
	// always send 0 for this. When this field is not 0, the LegacySetItemSlots slice below will have values
	// in it.
	// LegacyRequestID ties in with the ItemStackResponse packet. If this field is non-0, the server should
	// respond with an ItemStackResponse packet. Some inventory actions such as dropping an item out of the
	// hotbar are still one using this packet, and the ItemStackResponse packet needs to tie in with it.
	LegacyRequestID int32
	// LegacySetItemSlots are only present if the LegacyRequestID is non-zero. These item slots inform the
	// server of the slots that were changed during the inventory transaction, and the server should send
	// back an ItemStackResponse packet with these slots present in it. (Or false with no slots, if rejected.)
	LegacySetItemSlots []LegacySetItemSlot
	// Actions is a list of actions that took place, that form the inventory transaction together. Each of
	// these actions hold one slot in which one item was changed to another. In general, the combination of
	// all of these actions results in a balanced inventory transaction. This should be checked to ensure that
	// no items are cheated into the inventory.
	Actions []InventoryAction
	// ActionType is the type of the UseItem inventory transaction. It is one of the action types found above,
	// and specifies the way the player interacted with the block.
	ActionType uint32
	// TriggerType is the type of the trigger that caused the inventory transaction. It is one of the trigger
	// types found in the constants above. If TriggerType is TriggerTypePlayerInput, the transaction is from
	// the initial input of the player. If it is TriggerTypeSimulationTick, the transaction is from a simulation
	// tick when the player is holding down the input.
	TriggerType uint32
	// BlockPosition is the position of the block that was interacted with. This is only really a correct
	// block position if ActionType is not UseItemActionClickAir.
	BlockPosition BlockPos
	// BlockFace is the face of the block that was interacted with. When clicking the block, it is the face
	// clicked. When breaking the block, it is the face that was last being hit until the block broke.
	BlockFace int32
	// HotBarSlot is the hot bar slot that the player was holding while clicking the block. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was held to interact with the block. The server should check if this item
	// is actually present in the HotBarSlot.
	HeldItem ItemInstance
	// Position is the position of the player at the time of interaction. For clicking a block, this is the
	// position at that time, whereas for breaking the block it is the position at the time of breaking.
	Position mgl32.Vec3
	// ClickedPosition is the position that was clicked relative to the block's base coordinate. It can be
	// used to find out exactly where a player clicked the block.
	ClickedPosition mgl32.Vec3
	// BlockRuntimeID is the runtime ID of the block that was clicked. It may be used by the server to verify
	// that the player's world client-side is synchronised with the server's.
	BlockRuntimeID uint32
	// ClientPrediction is the client's prediction on the output of the transaction. It is one of the client
	// prediction found in the constants above.
	ClientPrediction byte
}

const (
	UseItemOnEntityActionInteract = iota
	UseItemOnEntityActionAttack
)

// UseItemOnEntityTransactionData represents an inventory transaction data object sent when the client uses
// an item on an entity.
type UseItemOnEntityTransactionData struct {
	// TargetEntityRuntimeID is the entity runtime ID of the target that was clicked. It is the runtime ID
	// that was assigned to it in the AddEntity packet.
	TargetEntityRuntimeID uint64
	// ActionType is the type of the UseItemOnEntity inventory transaction. It is one of the action types
	// found in the constants above, and specifies the way the player interacted with the entity.
	ActionType uint32
	// HotBarSlot is the hot bar slot that the player was holding while clicking the entity. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was held to interact with the entity. The server should check if this item
	// is actually present in the HotBarSlot.
	HeldItem ItemInstance
	// Position is the position of the player at the time of clicking the entity.
	Position mgl32.Vec3
	// ClickedPosition is the position that was clicked relative to the entity's base coordinate. It can be
	// used to find out exactly where a player clicked the entity.
	ClickedPosition mgl32.Vec3
}

const (
	ReleaseItemActionRelease = iota
	ReleaseItemActionConsume
)

// ReleaseItemTransactionData represents an inventory transaction data object sent when the client releases
// the item it was using, for example when stopping while eating or stopping the charging of a bow.
type ReleaseItemTransactionData struct {
	// ActionType is the type of the ReleaseItem inventory transaction. It is one of the action types found
	// in the constants above, and specifies the way the item was released.
	// As of 1.13, the ActionType is always 0. This field can be ignored, because releasing food (by consuming
	// it) or releasing a bow (to shoot an arrow) is essentially the same.
	ActionType uint32
	// HotBarSlot is the hot bar slot that the player was holding while releasing the item. It should be used
	// to ensure that the hot bar slot and held item are correctly synchronised with the server.
	HotBarSlot int32
	// HeldItem is the item that was released. The server should check if this item is actually present in the
	// HotBarSlot.
	HeldItem ItemInstance
	// HeadPosition is the position of the player's head at the time of releasing the item. This is used
	// mainly for purposes such as spawning eating particles at that position.
	HeadPosition mgl32.Vec3
}

// Marshal ...
func (data *UseItemTransactionData) Marshal(r IO) {
	r.Varuint32(&data.ActionType)
	r.Varuint32(&data.TriggerType)
	r.UBlockPos(&data.BlockPosition)
	r.Varint32(&data.BlockFace)
	r.Varint32(&data.HotBarSlot)
	r.ItemInstance(&data.HeldItem)
	r.Vec3(&data.Position)
	r.Vec3(&data.ClickedPosition)
	r.Varuint32(&data.BlockRuntimeID)
	r.Uint8(&data.ClientPrediction)
}

// Marshal ...
func (data *UseItemOnEntityTransactionData) Marshal(r IO) {
	r.Varuint64(&data.TargetEntityRuntimeID)
	r.Varuint32(&data.ActionType)
	r.Varint32(&data.HotBarSlot)
	r.ItemInstance(&data.HeldItem)
	r.Vec3(&data.Position)
	r.Vec3(&data.ClickedPosition)
}

// Marshal ...
func (data *ReleaseItemTransactionData) Marshal(r IO) {
	r.Varuint32(&data.ActionType)
	r.Varint32(&data.HotBarSlot)
	r.ItemInstance(&data.HeldItem)
	r.Vec3(&data.HeadPosition)
}

// Marshal ...
func (*NormalTransactionData) Marshal(IO) {}

// Marshal ...
func (*MismatchTransactionData) Marshal(IO) {}

// LegacySetItemSlot represents a slot that was changed during an InventoryTransaction. These slots have to
// have their values set accordingly for actions such as when dropping an item out of the hotbar, where the
// inventory container and the slot that had its item dropped is passed.
type LegacySetItemSlot struct {
	ContainerID byte
	Slots       []byte
}

// Marshal encodes/decodes a LegacySetItemSlot.
func (x *LegacySetItemSlot) Marshal(r IO) {
	r.Uint8(&x.ContainerID)
	r.ByteSlice(&x.Slots)
}
