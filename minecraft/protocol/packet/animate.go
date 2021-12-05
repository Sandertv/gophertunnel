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
	AnimateActionRowRight = iota + 128
	AnimateActionRowLeft
)

// Animate is sent by the server to send a player animation from one player to all viewers of that player. It
// is used for a couple of actions, such as arm swimming and critical hits.
type Animate struct {
	// ActionType is the ID of the animation action to execute. It is one of the action type constants that
	// may be found above.
	ActionType int32
	// EntityRuntimeID is the runtime ID of the player that the animation should be played upon. The runtime
	// ID is unique for each world session, and entities are generally identified in packets using this
	// runtime ID.
	EntityRuntimeID uint64
	// BoatRowingTime ...
	BoatRowingTime float32
}

// ID ...
func (*Animate) ID() uint32 {
	return IDAnimate
}

// Marshal ...
func (pk *Animate) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.ActionType)
	w.Varuint64(&pk.EntityRuntimeID)
	if pk.ActionType&0x80 != 0 {
		w.Float32(&pk.BoatRowingTime)
	}
}

// Unmarshal ...
func (pk *Animate) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.ActionType)
	r.Varuint64(&pk.EntityRuntimeID)
	if pk.ActionType&0x80 != 0 {
		r.Float32(&pk.BoatRowingTime)
	}
}
