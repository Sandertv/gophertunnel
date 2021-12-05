package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientBoundDebugRendererClear int32 = iota + 1
	ClientBoundDebugRendererAddCube
)

// ClientBoundDebugRenderer is sent by the server to spawn an outlined cube on client-side.
type ClientBoundDebugRenderer struct {
	// Type is the type of action. It is one of the constants above.
	Type int32
	// Text is the text that is displayed above the debug.
	Text string
	// Position is the position to spawn the debug on.
	Position mgl32.Vec3
	// Red is the red value from the RGBA colour rendered on the debug. This value is in the range 0-1.
	Red float32
	// Green is the green value from the RGBA colour rendered on the debug. This value is in the range 0-1.
	Green float32
	// Blue is the blue value from the RGBA colour rendered on the debug. This value is in the range 0-1.
	Blue float32
	// Alpha is the alpha value from the RGBA colour rendered on the debug. This value is in the range 0-1.
	Alpha float32
	// Duration is how long the debug will last in the world for. It is measured in milliseconds.
	Duration int64
}

// ID ...
func (*ClientBoundDebugRenderer) ID() uint32 {
	return IDClientBoundDebugRenderer
}

// Marshal ...
func (pk *ClientBoundDebugRenderer) Marshal(w *protocol.Writer) {
	w.Int32(&pk.Type)

	if pk.Type == ClientBoundDebugRendererAddCube {
		w.String(&pk.Text)
		w.Vec3(&pk.Position)
		w.Float32(&pk.Red)
		w.Float32(&pk.Green)
		w.Float32(&pk.Blue)
		w.Float32(&pk.Alpha)
		w.Int64(&pk.Duration)
	}
}

// Unmarshal ...
func (pk *ClientBoundDebugRenderer) Unmarshal(r *protocol.Reader) {
	r.Int32(&pk.Type)

	if pk.Type == ClientBoundDebugRendererAddCube {
		r.String(&pk.Text)
		r.Vec3(&pk.Position)
		r.Float32(&pk.Red)
		r.Float32(&pk.Green)
		r.Float32(&pk.Blue)
		r.Float32(&pk.Alpha)
		r.Int64(&pk.Duration)
	}
}
