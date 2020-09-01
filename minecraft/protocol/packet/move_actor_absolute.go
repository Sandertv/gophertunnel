package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	MoveFlagOnGround = 1 << iota
	MoveFlagTeleport
)

// MoveActorAbsolute is sent by the server to move an entity to an absolute position. It is typically used
// for movements where high accuracy isn't needed, such as for long range teleporting.
type MoveActorAbsolute struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	EntityRuntimeID uint64
	// Flags is a combination of flags that specify details of the movement. It is a combination of the flags
	// above.
	Flags byte
	// Position is the position to spawn the entity on. If the entity is on a distance that the player cannot
	// see it, the entity will still show up if the player moves closer.
	Position mgl32.Vec3
	// Rotation is a Vec3 holding the X, Y and Z rotation of the entity after the movement. This is a Vec3 for
	// the reason that projectiles like arrows don't have yaw/pitch, but do have roll.
	Rotation mgl32.Vec3
}

// ID ...
func (*MoveActorAbsolute) ID() uint32 {
	return IDMoveActorAbsolute
}

// Marshal ...
func (pk *MoveActorAbsolute) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Uint8(&pk.Flags)
	w.Vec3(&pk.Position)
	w.ByteFloat(&pk.Rotation[0])
	w.ByteFloat(&pk.Rotation[1])
	w.ByteFloat(&pk.Rotation[2])
}

// Unmarshal ...
func (pk *MoveActorAbsolute) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Uint8(&pk.Flags)
	r.Vec3(&pk.Position)
	r.ByteFloat(&pk.Rotation[0])
	r.ByteFloat(&pk.Rotation[1])
	r.ByteFloat(&pk.Rotation[2])
}
