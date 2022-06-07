package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PlayerAction is sent by the client when it executes any action, for example starting to sprint, swim,
// starting the breaking of a block, dropping an item, etc.
type PlayerAction struct {
	// EntityRuntimeID is the runtime ID of the player. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// ActionType is the ID of the action that was executed by the player. It is one of the constants that may
	// be found in protocol/player.go.
	ActionType int32
	// BlockPosition is the position of the target block, if the action with the ActionType set concerned a
	// block. If that is not the case, the block position will be zero.
	BlockPosition protocol.BlockPos
	// ResultPosition is the position of the action's result. When a UseItemOn action is sent, this is the position of
	// the block clicked, but when a block is placed, this is the position at which the block will be placed.
	ResultPosition protocol.BlockPos
	// BlockFace is the face of the target block that was touched. If the action with the ActionType set
	// concerned a block. If not, the face is always 0.
	BlockFace int32
}

// ID ...
func (*PlayerAction) ID() uint32 {
	return IDPlayerAction
}

// Marshal ...
func (pk *PlayerAction) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Varint32(&pk.ActionType)
	w.UBlockPos(&pk.BlockPosition)
	w.UBlockPos(&pk.ResultPosition)
	w.Varint32(&pk.BlockFace)
}

// Unmarshal ...
func (pk *PlayerAction) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Varint32(&pk.ActionType)
	r.UBlockPos(&pk.BlockPosition)
	r.UBlockPos(&pk.ResultPosition)
	r.Varint32(&pk.BlockFace)
}
