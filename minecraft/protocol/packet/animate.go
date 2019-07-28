package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	AnimateActionSwingArm = iota + 1
	_
	AnimateActionStopSleep
	AnimateActionCriticalHit
)

// Animate is sent by the server to send a player animation from one player to all viewers of that player. It
// is used for a couple of actions, such as arm swimming and critical hits.
type Animate struct {
	// ActionType is the ID of the animation action to execute. It is one of the action type constants that
	// may be found above.
	ActionType int32
	// EntityNetworkID is the runtime ID of the player that the animation should be played upon. The runtime
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
func (pk *Animate) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.ActionType)
	_ = protocol.WriteVaruint64(buf, pk.EntityRuntimeID)
	if pk.ActionType&0x80 != 0 {
		_ = protocol.WriteFloat32(buf, pk.BoatRowingTime)
	}
}

// Unmarshal ...
func (pk *Animate) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Varint32(buf, &pk.ActionType),
		protocol.Varuint64(buf, &pk.EntityRuntimeID),
	); err != nil {
		return err
	}
	if pk.ActionType&0x80 != 0 {
		return protocol.Float32(buf, &pk.BoatRowingTime)
	}
	return nil
}
