package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type CameraShake struct {
	// Intensity is the intensity of the shaking. The vanilla server limits this to 4, so a value larger
	// than 4 may not work.
	Intensity float32
	// Duration is the duration the camera will shake for. The unit of time used is currently unknown.
	Duration float32
}

// ID ...
func (CameraShake) ID() uint32 {
	return IDCameraShake
}

// Marshal ...
func (pk CameraShake) Marshal(w *protocol.Writer) {
	w.Float32(&pk.Intensity)
	w.Float32(&pk.Duration)
}

// Unmarshal ...
func (pk CameraShake) Unmarshal(r *protocol.Reader) {
	r.Float32(&pk.Intensity)
	r.Float32(&pk.Duration)
}
