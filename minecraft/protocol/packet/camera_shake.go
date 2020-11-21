package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	CameraShakeTypePositional uint8 = iota
	CameraShakeTypeRotational
)

// CameraShake is sent by the server to make the camera shake client-side. This feature was added for map-
// making partners.
type CameraShake struct {
	// Intensity is the intensity of the shaking. The client limits this value to 4, so anything higher may
	// not work.
	Intensity float32
	// Duration is the number of seconds the camera will shake for.
	Duration float32
	// Type is the type of shake, and is one of the constants listed above. The different type affects how
	// the shake looks in game.
	Type uint8
}

// ID ...
func (*CameraShake) ID() uint32 {
	return IDCameraShake
}

// Marshal ...
func (pk *CameraShake) Marshal(w *protocol.Writer) {
	w.Float32(&pk.Intensity)
	w.Float32(&pk.Duration)
	w.Uint8(&pk.Type)
}

// Unmarshal ...
func (pk *CameraShake) Unmarshal(r *protocol.Reader) {
	r.Float32(&pk.Intensity)
	r.Float32(&pk.Duration)
	r.Uint8(&pk.Type)
}
