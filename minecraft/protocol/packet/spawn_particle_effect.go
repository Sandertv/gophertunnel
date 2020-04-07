package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SpawnParticleEffect is sent by the server to spawn a particle effect client-side. Unlike other packets that
// result in the appearing of particles, this packet can show particles that are not hardcoded in the client.
// They can be added and changed through behaviour packs to implement custom particles.
type SpawnParticleEffect struct {
	// Dimension is the dimension that the particle is spawned in. Its exact usage is not clear, as the
	// dimension has no direct effect on the particle.
	Dimension byte
	// EntityUniqueID is the unique ID of the entity that the spawned particle may be attached to. If this ID
	// is not -1, the Position below will be interpreted as relative to the position of the entity associated
	// with this unique ID.
	EntityUniqueID int64
	// Position is the position that the particle should be spawned at. If the position is too far away from
	// the player, it will not show up.
	// If EntityUniqueID is not -1, the position will be relative to the position of the entity.
	Position mgl32.Vec3
	// ParticleName is the name of the particle that should be shown. This name may point to a particle effect
	// that is built-in, or to one implemented by behaviour packs.
	ParticleName string
}

// ID ...
func (*SpawnParticleEffect) ID() uint32 {
	return IDSpawnParticleEffect
}

// Marshal ...
func (pk *SpawnParticleEffect) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Dimension)
	_ = protocol.WriteVarint64(buf, pk.EntityUniqueID)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteString(buf, pk.ParticleName)
}

// Unmarshal ...
func (pk *SpawnParticleEffect) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.Dimension),
		protocol.Varint64(buf, &pk.EntityUniqueID),
		protocol.Vec3(buf, &pk.Position),
		protocol.String(buf, &pk.ParticleName),
	)
}
