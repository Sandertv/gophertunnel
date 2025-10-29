package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ScriptDebugShapeLine = iota
	ScriptDebugShapeBox
	ScriptDebugShapeSphere
	ScriptDebugShapeCircle
	ScriptDebugShapeText
	ScriptDebugShapeArrow
)

// DebugDrawerShape defines a single debug shape to be rendered on the client.
// Each shape has a unique NetworkID and a set of optional parameters depending on its type.
type DebugDrawerShape struct {
	// NetworkID is the network ID of the shape.
	NetworkID uint64
	// Type is the type of the shape.
	Type protocol.Optional[uint8]
	// Location is the location of the shape.
	Location protocol.Optional[mgl32.Vec3]
	// Scale is the scale of the shape.
	Scale protocol.Optional[float32]
	// Rotation is the rotation of the shape.
	Rotation protocol.Optional[mgl32.Vec3]
	// TotalTimeLeft is the total time left of the shape.
	TotalTimeLeft protocol.Optional[float32]
	// Colour is the ARGB colour of the shape.
	Colour protocol.Optional[color.RGBA]
	// Text is the text of the shape.
	Text protocol.Optional[string]
	// BoxBound is the box bound of the shape.
	BoxBound protocol.Optional[mgl32.Vec3]
	// LineEndLocation is the line end location of the shape.
	LineEndLocation protocol.Optional[mgl32.Vec3]
	// ArrowHeadLength is the arrow head length of the shape.
	ArrowHeadLength protocol.Optional[float32]
	// ArrowHeadRadius is the arrow head radius of the shape.
	ArrowHeadRadius protocol.Optional[float32]
	// Segments is the number of segments of the shape.
	Segments protocol.Optional[byte]
}

func (x *DebugDrawerShape) Marshal(io protocol.IO) {
	io.Varuint64(&x.NetworkID)
	protocol.OptionalFunc(io, &x.Type, io.Uint8)
	protocol.OptionalFunc(io, &x.Location, io.Vec3)
	protocol.OptionalFunc(io, &x.Scale, io.Float32)
	protocol.OptionalFunc(io, &x.Rotation, io.Vec3)
	protocol.OptionalFunc(io, &x.TotalTimeLeft, io.Float32)
	protocol.OptionalFunc(io, &x.Colour, io.BEARGB)
	protocol.OptionalFunc(io, &x.Text, io.String)
	protocol.OptionalFunc(io, &x.BoxBound, io.Vec3)
	protocol.OptionalFunc(io, &x.LineEndLocation, io.Vec3)
	protocol.OptionalFunc(io, &x.ArrowHeadLength, io.Float32)
	protocol.OptionalFunc(io, &x.ArrowHeadRadius, io.Float32)
	protocol.OptionalFunc(io, &x.Segments, io.Uint8)
}

// DebugDrawer is a packet sent by the server to instruct the client to render
// one or more debug shapes. Each packet fully replaces any previously rendered shapes.
// To remove a shape, omit it from the next packet's Shapes slice.
type DebugDrawer struct {
	// Shapes is a list of shapes to draw on the client-side.
	Shapes []DebugDrawerShape
}

// ID ...
func (pk *DebugDrawer) ID() uint32 {
	return IDDebugDrawer
}

func (pk *DebugDrawer) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Shapes)
}
