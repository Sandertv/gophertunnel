package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientInputLockCamera = 1 << (iota + 1)
	ClientInputLockMovement
	_
	ClientInputLockLateralMovement
	ClientInputLockSneak
	ClientInputLockJump
	ClientInputLockMount
	ClientInputLockDismount
	ClientInputLockMoveForward
	ClientInputLockMoveBackward
	ClientInputLockMoveLeft
	ClientInputLockMoveRight
)

// UpdateClientInputLocks is sent by the server to the client to lock specific player inputs such as camera
// rotation, movement, jumping, sneaking, mounting or individual directional movement.
type UpdateClientInputLocks struct {
	// Locks is a set of flags that specify which client inputs are disabled, such as whether the player can
	// move, rotate the camera, jump, sneak or mount/dismount entities. It is a combination of the
	// ClientInputLock constants above.
	Locks uint32
	// Position is the server's position of the client at the time the packet was sent. It is unclear what the exact
	// purpose of this field is.
	Position mgl32.Vec3
}

// ID ...
func (pk *UpdateClientInputLocks) ID() uint32 {
	return IDUpdateClientInputLocks
}

func (pk *UpdateClientInputLocks) Marshal(io protocol.IO) {
	io.Varuint32(&pk.Locks)
	io.Vec3(&pk.Position)
}
