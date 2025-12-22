package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AnimateActionSwingArm = iota + 1
	_
	AnimateActionStopSleep
	AnimateActionCriticalHit
	AnimateActionMagicCriticalHit
)

const (
	AnimateSwingSourceNone      = "none"
	AnimateSwingSourceBuild     = "build"
	AnimateSwingSourceMine      = "mine"
	AnimateSwingSourceInteract  = "interact"
	AnimateSwingSourceAttack    = "attack"
	AnimateSwingSourceUseItem   = "useitem"
	AnimateSwingSourceThrowItem = "throwitem"
	AnimateSwingSourceDropItem  = "dropitem"
	AnimateSwingSourceEvent     = "event"
)

// Animate is sent by the server to send a player animation from one player to all viewers of that player. It
// is used for a couple of actions, such as arm swimming and critical hits.
type Animate struct {
	// ActionType is the ID of the animation action to execute. It is one of the action type constants that
	// may be found above.
	ActionType uint8
	// EntityRuntimeID is the runtime ID of the player that the animation should be played upon. The runtime
	// ID is unique for each world session, and entities are generally identified in packets using this
	// runtime ID.
	EntityRuntimeID uint64
	// Data ...
	Data float32
	// SwingSource is the source for swing actions. It is one of the action type constants that
	// may be found above.
	SwingSource protocol.Optional[string]
}

// ID ...
func (*Animate) ID() uint32 {
	return IDAnimate
}

func (pk *Animate) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionType)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Float32(&pk.Data)
	protocol.OptionalFunc(io, &pk.SwingSource, io.String)
}
