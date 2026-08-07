package packet

import (
	"image/color"

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
	// Data holds the marker data. It is only present when Type is
	// ClientBoundDebugRendererAddCube.
	Data protocol.Optional[DebugMarkerData]
}

// DebugMarkerData holds the data of a debug marker cube spawned through a
// ClientBoundDebugRenderer packet.
type DebugMarkerData struct {
	// Text is the text that is displayed above the debug.
	Text string
	// Position is the position to spawn the debug on.
	Position mgl32.Vec3
	// Colour is the RGBA colour rendered on the debug, packed as an ARGB uint32
	// on the wire.
	Colour color.RGBA
	// Duration is how long the debug will last in the world for. It is measured in milliseconds.
	Duration uint64
}

// Marshal ...
func (x *DebugMarkerData) Marshal(io protocol.IO) {
	io.String(&x.Text)
	io.Vec3(&x.Position)
	io.ARGB(&x.Colour)
	io.Uint64(&x.Duration)
}

// ID ...
func (*ClientBoundDebugRenderer) ID() uint32 {
	return IDClientBoundDebugRenderer
}

func (pk *ClientBoundDebugRenderer) Marshal(io protocol.IO) {
	typStr := clientBoundDebugRenderToString(pk.Type)
	io.String(&typStr)
	clientBoundDebugRenderFromString(io, &pk.Type, typStr)
	protocol.OptionalMarshaler(io, &pk.Data)
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

func clientBoundDebugRenderFromString(io protocol.IO, x *uint32, s string) {
	switch s {
	case "cleardebugmarkers":
		*x = ClientBoundDebugRendererClear
	case "adddebugmarkercube":
		*x = ClientBoundDebugRendererAddCube
	default:
		io.InvalidValue(s, "type", "unknown type")
	}
}
