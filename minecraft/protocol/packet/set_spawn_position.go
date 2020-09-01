package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SpawnTypePlayer = iota
	SpawnTypeWorld
)

// SetSpawnPosition is sent by the server to update the spawn position of a player, for example when sleeping
// in a bed.
type SetSpawnPosition struct {
	// SpawnType is the type of spawn to set. It is either SpawnTypePlayer or SpawnTypeWorld, and specifies
	// the behaviour of the spawn set. If SpawnTypeWorld is set, the position to which compasses will point is
	// also changed.
	SpawnType int32
	// Position is the new position of the spawn that was set. If SpawnType is SpawnTypeWorld, compasses will
	// point to this position. As of 1.16, Position is always the position of the player.
	Position protocol.BlockPos
	// Dimension is the ID of the dimension that had its spawn updated. This is specifically relevant for
	// behaviour added in 1.16 such as the respawn anchor, which allows setting the spawn in a specific
	// dimension.
	Dimension int32
	// SpawnPosition is a new field added in 1.16. It holds the spawn position of the world. This spawn
	// position is {-2147483648, -2147483648, -2147483648} for a default spawn position.
	SpawnPosition protocol.BlockPos
}

// ID ...
func (*SetSpawnPosition) ID() uint32 {
	return IDSetSpawnPosition
}

// Marshal ...
func (pk *SetSpawnPosition) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.SpawnType)
	w.UBlockPos(&pk.Position)
	w.Varint32(&pk.Dimension)
	w.UBlockPos(&pk.SpawnPosition)
}

// Unmarshal ...
func (pk *SetSpawnPosition) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.SpawnType)
	r.UBlockPos(&pk.Position)
	r.Varint32(&pk.Dimension)
	r.UBlockPos(&pk.SpawnPosition)
}
