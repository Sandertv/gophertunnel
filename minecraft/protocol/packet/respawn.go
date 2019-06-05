package packet

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Respawn is sent by the server to make a player respawn client-side. It is sent in response to a
// PlayerAction packet with ActionType PlayerActionRespawn.
type Respawn struct {
	// Position is the position on which the player should be respawned. The position might be in a different
	// dimension, in which case the client should first be sent a ChangeDimension packet.
	Position mgl32.Vec3
}

// ID ...
func (*Respawn) ID() uint32 {
	return IDRespawn
}

// Marshal ...
func (pk *Respawn) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVec3(buf, pk.Position)
}

// Unmarshal ...
func (pk *Respawn) Unmarshal(buf *bytes.Buffer) error {
	return protocol.Vec3(buf, &pk.Position)
}
