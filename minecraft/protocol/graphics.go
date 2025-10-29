package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

const (
	GraphicsOverrideParameterTypeSkyZenithColor uint8 = iota
)

// ParameterKeyframeValue represents a keyframe value for graphics parameters.
type ParameterKeyframeValue struct {
	// Time is the time for this keyframe.
	Time float32
	// Value is the value at this keyframe.
	Value mgl32.Vec3
}

// Marshal encodes/decodes a ParameterKeyframeValue.
func (x *ParameterKeyframeValue) Marshal(r IO) {
	r.Float32(&x.Time)
	r.Vec3(&x.Value)
}
