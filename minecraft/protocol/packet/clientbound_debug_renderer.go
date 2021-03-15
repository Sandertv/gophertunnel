package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientBoundDebugRendererTypeClear int32 = iota + 1
	ClientBoundDebugRendererTypeAddCube
)

// ClientBoundDebugRenderer
type ClientBoundDebugRenderer struct {
	// Type ...
	Type int32
	// Text ...
	Text string
	// Position ...
	Position mgl32.Vec3
	// Red ...
	Red float32
	// Green ...
	Green float32
	// Blue ...
	Blue float32
	// Alpha ...
	Alpha float32
	// Duration ...
	Duration int64
}

// ID ...
func (*ClientBoundDebugRenderer) ID() uint32 {
	return IDClientBoundDebugRenderer
}

// Marshal ...
func (pk *ClientBoundDebugRenderer) Marshal(w *protocol.Writer) {
	w.Int32(&pk.Type)

	switch pk.Type {
	case ClientBoundDebugRendererTypeClear:
	case ClientBoundDebugRendererTypeAddCube:
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

	switch pk.Type {
	case ClientBoundDebugRendererTypeClear:
	case ClientBoundDebugRendererTypeAddCube:
		r.String(&pk.Text)
		r.Vec3(&pk.Position)
		r.Float32(&pk.Red)
		r.Float32(&pk.Green)
		r.Float32(&pk.Blue)
		r.Float32(&pk.Alpha)
		r.Int64(&pk.Duration)
	}
}
