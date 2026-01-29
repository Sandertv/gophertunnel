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
	AnimateSwingSourceNone = iota + 1
	AnimateSwingSourceBuild
	AnimateSwingSourceMine
	AnimateSwingSourceInteract
	AnimateSwingSourceAttack
	AnimateSwingSourceUseItem
	AnimateSwingSourceThrowItem
	AnimateSwingSourceDropItem
	AnimateSwingSourceEvent
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
	SwingSource uint8
}

// ID ...
func (*Animate) ID() uint32 {
	return IDAnimate
}

func (pk *Animate) Marshal(io protocol.IO) {
	var swingSource protocol.Optional[string]
	if pk.SwingSource != 0 {
		swingSource = protocol.Option(swingSourceToString(pk.SwingSource))
	}
	io.Uint8(&pk.ActionType)
	io.Varuint64(&pk.EntityRuntimeID)
	io.Float32(&pk.Data)
	protocol.OptionalFunc(io, &swingSource, io.String)
	if val, ok := swingSource.Value(); ok {
		swingSourceFromString(io, &pk.SwingSource, val)
	}
}

func swingSourceFromString(io protocol.IO, x *uint8, s string) {
	switch s {
	case "none":
		*x = AnimateSwingSourceNone
	case "build":
		*x = AnimateSwingSourceBuild
	case "mine":
		*x = AnimateSwingSourceMine
	case "interact":
		*x = AnimateSwingSourceInteract
	case "attack":
		*x = AnimateSwingSourceAttack
	case "useitem":
		*x = AnimateSwingSourceUseItem
	case "throwitem":
		*x = AnimateSwingSourceThrowItem
	case "dropitem":
		*x = AnimateSwingSourceDropItem
	case "event":
		*x = AnimateSwingSourceEvent
	default:
		io.InvalidValue(s, "swingSource", "unknown source")
	}
}

func swingSourceToString(x uint8) string {
	switch x {
	case AnimateSwingSourceNone:
		return "none"
	case AnimateSwingSourceBuild:
		return "build"
	case AnimateSwingSourceMine:
		return "mine"
	case AnimateSwingSourceInteract:
		return "interact"
	case AnimateSwingSourceAttack:
		return "attack"
	case AnimateSwingSourceUseItem:
		return "useitem"
	case AnimateSwingSourceThrowItem:
		return "throwitem"
	case AnimateSwingSourceDropItem:
		return "dropitem"
	case AnimateSwingSourceEvent:
		return "event"
	default:
		return "unknown"
	}
}
