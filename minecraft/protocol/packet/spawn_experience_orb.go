package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SpawnExperienceOrb is sent by the server to spawn an experience orb entity client-side. Much like the
// AddPainting packet, it is one of the few packets that spawn an entity without using the AddActor packet.
type SpawnExperienceOrb struct {
	// Position is the position to spawn the experience orb on. If the entity is on a distance that the player
	// cannot see it, the entity will still show up if the player moves closer.
	Position mgl32.Vec3
	// ExperienceAmount is the amount of experience in experience points that the orb carries. The client-side
	// size of the orb depends on the amount of experience in the orb: There are 11 possible sizes for the
	// orb, for 1–2, 3–6, 7–16, 17–36, 37–72, 73–148, 149–306, 307–616, 617–1236, 1237–2476, and 2477 and up.
	ExperienceAmount int32
}

// ID ...
func (*SpawnExperienceOrb) ID() uint32 {
	return IDSpawnExperienceOrb
}

// Marshal ...
func (pk *SpawnExperienceOrb) Marshal(w *protocol.Writer) {
	w.Vec3(&pk.Position)
	w.Varint32(&pk.ExperienceAmount)
}

// Unmarshal ...
func (pk *SpawnExperienceOrb) Unmarshal(r *protocol.Reader) {
	r.Vec3(&pk.Position)
	r.Varint32(&pk.ExperienceAmount)
}
