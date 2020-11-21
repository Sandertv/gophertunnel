package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type MotionPredictionHints struct {
	EntityRuntimeID uint64
	Motion          mgl32.Vec3
	OnGround        bool
}

// ID ...
func (*MotionPredictionHints) ID() uint32 {
	return IDMotionPredictionHints
}

// Marshal ...
func (pk *MotionPredictionHints) Marshal(w *protocol.Writer) {
	w.Varuint64(&pk.EntityRuntimeID)
	w.Vec3(&pk.Motion)
	w.Bool(&pk.OnGround)
}

// Unmarshal ...
func (pk *MotionPredictionHints) Unmarshal(r *protocol.Reader) {
	r.Varuint64(&pk.EntityRuntimeID)
	r.Vec3(&pk.Motion)
	r.Bool(&pk.OnGround)
}
