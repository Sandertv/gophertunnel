package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	DimensionOverworld = iota
	DimensionNether
	DimensionEnd
)

// ChangeDimension is sent by the server to the client to send a dimension change screen client-side. Once the
// screen is cleared client-side, the client will send a PlayerAction packet with
// PlayerActionDimensionChangeDone.
type ChangeDimension struct {
	// Dimension is the dimension that the client should be changed to. The fog colour will change depending
	// on the type of dimension, which is one of the constants above.
	// Note that Dimension MUST be a different dimension than the one that the player is currently in. Sending
	// a ChangeDimension packet with a Dimension that the player is currently in will result in a never-ending
	// dimension change screen.
	Dimension int32
	// Position is the position in the new dimension that the player is spawned in.
	Position mgl32.Vec3
	// Respawn specifies if the dimension change was respawn based, meaning that the player died in one
	// dimension and got respawned into another. The client will send a PlayerAction packet with
	// PlayerActionDimensionChangeRequest if it dies in another dimension, indicating that it needs a
	// DimensionChange packet with Respawn set to true.
	Respawn bool
}

// ID ...
func (*ChangeDimension) ID() uint32 {
	return IDChangeDimension
}

// Marshal ...
func (pk *ChangeDimension) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.Dimension)
	w.Vec3(&pk.Position)
	w.Bool(&pk.Respawn)
}

// Unmarshal ...
func (pk *ChangeDimension) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.Dimension)
	r.Vec3(&pk.Position)
	r.Bool(&pk.Respawn)
}
