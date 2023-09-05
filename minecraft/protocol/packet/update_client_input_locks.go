package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientInputLockCamera = 1 << (iota + 1)
	ClientInputLockMovement
)

// UpdateClientInputLocks is sent by the server to the client to lock either the camera or physical movement of the client.
type UpdateClientInputLocks struct {
	// Locks is a bitset that controls which locks are active. It is a combination of the constants above. If the camera
	// is locked, then the player cannot change their pitch or yaw. If movement is locked, the player cannot move in any
	// direction, they cannot jump, sneak or mount/dismount from any entities.
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
