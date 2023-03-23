package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientInputLockMove = 1 << (iota + 1)
	ClientInputLockJump
	ClientInputLockSneak
	ClientInputLockMount
	ClientInputLockDismount
	ClientInputLockRotation
)

// UpdateClientInputLocks is sent by the server to the client to lock certain inputs the client usually has, such as
// movement, jumping, sneaking, and more.
type UpdateClientInputLocks struct {
	// Locks is an encoded bitset of all locks that are currently active. The locks are defined in the constants above.
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
