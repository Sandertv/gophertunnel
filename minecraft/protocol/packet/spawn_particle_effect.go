package packet

import (
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
	// MoLangVariables is an encoded JSON map of MoLang variables that may be applicable to the particle spawn. This can
	// just be left empty in most cases.
	MoLangVariables protocol.Optional[[]byte]
}

// ID ...
func (*SpawnParticleEffect) ID() uint32 {
	return IDSpawnParticleEffect
}

func (pk *SpawnParticleEffect) Marshal(io protocol.IO) {
	io.Uint8(&pk.Dimension)
	io.Varint64(&pk.EntityUniqueID)
	io.Vec3(&pk.Position)
	io.String(&pk.ParticleName)
	protocol.OptionalFunc(io, &pk.MoLangVariables, io.ByteSlice)
}
