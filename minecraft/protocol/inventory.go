package protocol

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	InventoryActionSourceContainer = 0
	InventoryActionSourceWorld     = 2
	InventoryActionSourceCreative  = 3
	InventoryActionSourceTODO      = 99999
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
	OldItem ItemStack
	// NewItem is the new item that was put in the InventorySlot that the OldItem was in. It must be checked
	// in combination with other inventory actions to ensure that the transaction is balanced.
	NewItem ItemStack
}

// InvAction reads an inventory action from buffer src.
func InvAction(src *bytes.Buffer, action *InventoryAction) error {
	if err := Varuint32(src, &action.SourceType); err != nil {
		return wrap(err)
	}
	switch action.SourceType {
	case InventoryActionSourceContainer, InventoryActionSourceCraftingGrid, InventoryActionSourceTODO:
		if err := Varint32(src, &action.WindowID); err != nil {
			return wrap(err)
		}
	case InventoryActionSourceWorld:
		if err := Varuint32(src, &action.SourceFlags); err != nil {
			return wrap(err)
		}
	}
	return chainErr(
		Varuint32(src, &action.InventorySlot),
		Item(src, &action.OldItem),
		Item(src, &action.NewItem),
	)
}

// WriteInvAction writes an inventory action to buffer dst.
func WriteInvAction(dst *bytes.Buffer, action InventoryAction) error {
	if err := WriteVaruint32(dst, action.SourceType); err != nil {
		return wrap(err)
	}
	switch action.SourceType {
	case InventoryActionSourceContainer, InventoryActionSourceCraftingGrid, InventoryActionSourceTODO:
		if err := WriteVarint32(dst, action.WindowID); err != nil {
			return wrap(err)
		}
	case InventoryActionSourceWorld:
		if err := WriteVaruint32(dst, action.SourceFlags); err != nil {
			return wrap(err)
		}
	}
	return chainErr(
		WriteVaruint32(dst, action.InventorySlot),
		WriteItem(dst, action.OldItem),
		WriteItem(dst, action.NewItem),
	)
}

// InventoryTransactionData represents an object that holds data specific to an inventory transaction type.
// The data it holds depends on the type.
type InventoryTransactionData interface {
	// Marshal encodes the inventory transaction data to its binary representation into buf.
	Marshal(buf *bytes.Buffer)
	// Unmarshal decodes a serialised inventory transaction data object in buf into the
	// InventoryTransactionData instance.
	Unmarshal(buf *bytes.Buffer) error
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

// UseItemTransactionData represents an inventory transaction data object sent when the client uses an item on
// a block.
type UseItemTransactionData struct {
	// ActionType is the type of the UseItem inventory transaction. It is one of the action types found above,
	// and specifies the way the player interacted with the block.
	ActionType uint32
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
	HeldItem ItemStack
	// Position is the position of the player at the time of interaction. For clicking a block, this is the
	// position at that time, whereas for breaking the block it is the position at the time of breaking.
	Position mgl32.Vec3
	// ClickedPosition is the position that was clicked relative to the block's base coordinate. It can be
	// used to find out exactly where a player clicked the block.
	ClickedPosition mgl32.Vec3
	// BlockRuntimeID is the runtime ID of the block that was clicked. It may be used by the server to verify
	// that the player's world client-side is synchronised with the server's.
	BlockRuntimeID uint32
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
	HeldItem ItemStack
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
	HeldItem ItemStack
	// HeadPosition is the position of the player's head at the time of releasing the item. This is used
	// mainly for purposes such as spawning eating particles at that position.
	HeadPosition mgl32.Vec3
}

// Marshal ...
func (data *UseItemTransactionData) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint32(buf, data.ActionType)
	_ = WriteUBlockPosition(buf, data.BlockPosition)
	_ = WriteVarint32(buf, data.BlockFace)
	_ = WriteVarint32(buf, data.HotBarSlot)
	_ = WriteItem(buf, data.HeldItem)
	_ = WriteVec3(buf, data.Position)
	_ = WriteVec3(buf, data.ClickedPosition)
	_ = WriteVaruint32(buf, data.BlockRuntimeID)
}

// Unmarshal ...
func (data *UseItemTransactionData) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		Varuint32(buf, &data.ActionType),
		UBlockPosition(buf, &data.BlockPosition),
		Varint32(buf, &data.BlockFace),
		Varint32(buf, &data.HotBarSlot),
		Item(buf, &data.HeldItem),
		Vec3(buf, &data.Position),
		Vec3(buf, &data.ClickedPosition),
		Varuint32(buf, &data.BlockRuntimeID),
	)
}

// Marshal ...
func (data *UseItemOnEntityTransactionData) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint64(buf, data.TargetEntityRuntimeID)
	_ = WriteVaruint32(buf, data.ActionType)
	_ = WriteVarint32(buf, data.HotBarSlot)
	_ = WriteItem(buf, data.HeldItem)
	_ = WriteVec3(buf, data.Position)
	_ = WriteVec3(buf, data.ClickedPosition)
}

// Unmarshal ...
func (data *UseItemOnEntityTransactionData) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		Varuint64(buf, &data.TargetEntityRuntimeID),
		Varuint32(buf, &data.ActionType),
		Varint32(buf, &data.HotBarSlot),
		Item(buf, &data.HeldItem),
		Vec3(buf, &data.Position),
		Vec3(buf, &data.ClickedPosition),
	)
}

// Marshal ...
func (data *ReleaseItemTransactionData) Marshal(buf *bytes.Buffer) {
	_ = WriteVaruint32(buf, data.ActionType)
	_ = WriteVarint32(buf, data.HotBarSlot)
	_ = WriteItem(buf, data.HeldItem)
	_ = WriteVec3(buf, data.HeadPosition)
}

// Unmarshal ...
func (data *ReleaseItemTransactionData) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		Varuint32(buf, &data.ActionType),
		Varint32(buf, &data.HotBarSlot),
		Item(buf, &data.HeldItem),
		Vec3(buf, &data.HeadPosition),
	)
}

// Marshal ...
func (*NormalTransactionData) Marshal(buf *bytes.Buffer) {
	// No payload.
}

// Unmarshal ...
func (*NormalTransactionData) Unmarshal(buf *bytes.Buffer) error {
	return nil
}

// Marshal ...
func (*MismatchTransactionData) Marshal(buf *bytes.Buffer) {
	// No payload.
}

// Unmarshal ...
func (*MismatchTransactionData) Unmarshal(buf *bytes.Buffer) error {
	return nil
}
