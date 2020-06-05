package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerActionStartBreak = iota
	PlayerActionAbortBreak
	PlayerActionStopBreak
	PlayerActionGetUpdatedBlock
	PlayerActionDropItem
	PlayerActionStartSleeping
	PlayerActionStopSleeping
	PlayerActionRespawn
	PlayerActionJump
	PlayerActionStartSprint
	PlayerActionStopSprint
	PlayerActionStartSneak
	PlayerActionStopSneak
	PlayerActionDimensionChangeRequest // Deprecated.
	PlayerActionDimensionChangeDone
	PlayerActionStartGlide
	PlayerActionStopGlide
	PlayerActionBuildDenied
	PlayerActionContinueBreak
	PlayerActionChangeSkin
	PlayerActionSetEnchantmentSeed
	PlayerActionStartSwimming
	PlayerActionStopSwimming
	PlayerActionStartSpinAttack
	PlayerActionStopSpinAttack
	PlayerActionStartBuildingBlock
)

// PlayerAction is sent by the client when it executes any action, for example starting to sprint, swim,
// starting the breaking of a block, dropping an item, etc.
type PlayerAction struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// ActionType is the ID of the action that was executed by the player. It is one of the constants that may
	// be found above.
	ActionType int32
	// BlockPosition is the position of the target block, if the action with the ActionType set concerned a
	// block. If that is not the case, the block position will be zero.
	BlockPosition protocol.BlockPos
	// BlockFace is the face of the target block that was touched. If the action with the ActionType set
	// concerned a block. If not, the face is always 0.
	BlockFace int32
}

// ID ...
func (*PlayerAction) ID() uint32 {
	return IDPlayerAction
}

// Marshal ...
func (pk *PlayerAction) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	_ = protocol.WriteVarint32(buf, pk.ActionType)
	_ = protocol.WriteUBlockPosition(buf, pk.BlockPosition)
	_ = protocol.WriteVarint32(buf, pk.BlockFace)
}

// Unmarshal ...
func (pk *PlayerAction) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
		protocol.Varint32(buf, &pk.ActionType),
		protocol.UBlockPosition(buf, &pk.BlockPosition),
		protocol.Varint32(buf, &pk.BlockFace),
	)
}
