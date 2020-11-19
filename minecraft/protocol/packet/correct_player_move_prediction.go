package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type CorrectPlayerMovePrediction struct {
	Position mgl32.Vec3
	Delta    mgl32.Vec3
	OnGround bool
	Tick     uint64
}

// ID ...
func (pk CorrectPlayerMovePrediction) ID() uint32 {
	return IDCorrectPlayerMovePrediction
}

// Marshal ...
func (pk CorrectPlayerMovePrediction) Marshal(w *protocol.Writer) {
	w.Vec3(&pk.Position)
	w.Vec3(&pk.Delta)
	w.Bool(&pk.OnGround)
	w.Varuint64(&pk.Tick)
}

// Unmarshal ...
func (pk CorrectPlayerMovePrediction) Unmarshal(r *protocol.Reader) {
	r.Vec3(&pk.Position)
	r.Vec3(&pk.Delta)
	r.Bool(&pk.OnGround)
	r.Varuint64(&pk.Tick)
}
