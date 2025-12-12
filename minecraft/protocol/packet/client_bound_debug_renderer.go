package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ClientBoundDebugRendererClear = iota
	ClientBoundDebugRendererAddCube
)

// ClientBoundDebugRenderer is sent by the server to spawn an outlined cube on client-side.
type ClientBoundDebugRenderer struct {
	// Type is the type of action. It is one of the constants above.
	Type uint32
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
	Duration uint64
}

// ID ...
func (*ClientBoundDebugRenderer) ID() uint32 {
	return IDClientBoundDebugRenderer
}

func (pk *ClientBoundDebugRenderer) Marshal(io protocol.IO) {
	typStr := clientBoundDebugRenderToString(pk.Type)
	io.String(&typStr)
	if typ, ok := clientBoundDebugRenderFromString(typStr); ok {
		pk.Type = typ
	} else {
		io.InvalidValue(typ, "type", "unknown type")
	}
	if pk.Type == ClientBoundDebugRendererAddCube {
		io.String(&pk.Text)
		io.Vec3(&pk.Position)
		io.Float32(&pk.Red)
		io.Float32(&pk.Green)
		io.Float32(&pk.Blue)
		io.Float32(&pk.Alpha)
		io.Uint64(&pk.Duration)
	}
}

func clientBoundDebugRenderToString(x uint32) string {
	switch x {
	case ClientBoundDebugRendererClear:
		return "cleardebugmarkers"
	case ClientBoundDebugRendererAddCube:
		return "adddebugmarkercube"
	default:
		return "unknown"
	}
}

func clientBoundDebugRenderFromString(x string) (uint32, bool) {
	switch x {
	case "cleardebugmarkers":
		return ClientBoundDebugRendererClear, true
	case "adddebugmarkercube":
		return ClientBoundDebugRendererAddCube, true
	default:
		return 0, false
	}
}
