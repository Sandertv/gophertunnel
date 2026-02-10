package protocol

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	ShapeDataLast = iota
	ShapeDataArrow
	ShapeDataText
	ShapeDataBox
	ShapeDataLine
	ShapeDataSphere
)

// lookupShapeData looks up an ShapeData matching the shape data type passed.
// False is returned if no such shape data exists.
func lookupShapeData(shapeDataType uint32, x *ShapeData) bool {
	switch shapeDataType {
	case ShapeDataLast:
		*x = &LastShape{}
	case ShapeDataArrow:
		*x = &ArrowShape{}
	case ShapeDataText:
		*x = &TextShape{}
	case ShapeDataBox:
		*x = &BoxShape{}
	case ShapeDataLine:
		*x = &LineShape{}
	case ShapeDataSphere:
		*x = &SphereShape{}
	default:
		return false
	}
	return true
}

// lookupShapeDataType looks up a debug shape type that matches the ShapeData passed.
func lookupShapeDataType(x ShapeData, shapeDataType *uint32) bool {
	switch x.(type) {
	case *LastShape:
		*shapeDataType = ShapeDataLast
	case *ArrowShape:
		*shapeDataType = ShapeDataArrow
	case *TextShape:
		*shapeDataType = ShapeDataText
	case *BoxShape:
		*shapeDataType = ShapeDataBox
	case *LineShape:
		*shapeDataType = ShapeDataLine
	case *SphereShape:
		*shapeDataType = ShapeDataSphere
	default:
		return false
	}
	return true
}

// ShapeData represents an object that holds data specific to a debug shape.
// The data it holds depends on the type.
type ShapeData interface {
	// Marshal encodes/decodes a serialised debug shape data object.
	Marshal(r IO)
}

// LastShape points to using the last shape settings.
// If no shape has ever been set, then use the default one.
type LastShape struct{}

// Marshal ...
func (shape *LastShape) Marshal(io IO) {}

// LineShape represents a line debug shape.
type LineShape struct {
	// LineEndLocation is the line end location of the shape.
	LineEndLocation mgl32.Vec3
}

// Marshal ...
func (shape *LineShape) Marshal(io IO) {
	io.Vec3(&shape.LineEndLocation)
}

// TextShape represents a text debug shape.
type TextShape struct {
	// Text is the text of the debug text shape.
	Text string
}

// Marshal ...
func (shape *TextShape) Marshal(io IO) {
	io.String(&shape.Text)
}

// BoxShape represents a box debug shape.
type BoxShape struct {
	// BoxBound is the box bound of the shape.
	BoxBound mgl32.Vec3
}

// Marshal ...
func (shape *BoxShape) Marshal(io IO) {
	io.Vec3(&shape.BoxBound)
}

// SphereShape represents a circle or sphere debug shape.
type SphereShape struct {
	// Segments is the segments that used for the debug circle or sphere.
	Segments byte
}

// Marshal ...
func (shape *SphereShape) Marshal(io IO) {
	io.Uint8(&shape.Segments)
}

// ArrowShape represents an arrow debug shape.
type ArrowShape struct {
	// ArrowEndLocation is the arrow end location of the shape.
	ArrowEndLocation Optional[mgl32.Vec3]
	// ArrowHeadLength is the arrow head length of the shape.
	ArrowHeadLength Optional[float32]
	// ArrowHeadRadius is the arrow head radius of the shape.
	ArrowHeadRadius Optional[float32]
	// Segments is the segments that used for the debug arrow's head.
	Segments Optional[byte]
}

// Marshal ...
func (shape *ArrowShape) Marshal(io IO) {
	OptionalFunc(io, &shape.ArrowEndLocation, io.Vec3)
	OptionalFunc(io, &shape.ArrowHeadLength, io.Float32)
	OptionalFunc(io, &shape.ArrowHeadRadius, io.Float32)
	OptionalFunc(io, &shape.Segments, io.Uint8)
}

const (
	DebugDrawerShapeLine = iota
	DebugDrawerShapeBox
	DebugDrawerShapeSphere
	DebugDrawerShapeCircle
	DebugDrawerShapeText
	DebugDrawerShapeArrow
)

// DebugDrawerShape defines a single debug shape to be rendered on the client.
// Each shape has a unique NetworkID and a set of optional parameters depending on its type.
type DebugDrawerShape struct {
	// NetworkID is the network ID of the shape.
	NetworkID uint64
	// DimensionID is the optional dimension ID where the shape is rendered.
	DimensionID Optional[int32]
	// AttachedToEntityID is the optional runtime ID of the entity the shape is attached to.
	AttachedToEntityID Optional[uint64]
	// Type is the type of the shape.
	// If not set, the set shape will be cleared.
	Type Optional[uint8]
	// Location is the location of the shape.
	Location Optional[mgl32.Vec3]
	// Scale is the scale of the shape.
	Scale Optional[float32]
	// Rotation is the rotation of the shape.
	Rotation Optional[mgl32.Vec3]
	// TotalTimeLeft is the total time left of the shape.
	TotalTimeLeft Optional[float32]
	// Colour is the ARGB colour of the shape.
	Colour Optional[color.RGBA]
	// ExtraShapeData holding data specific to the type of shape (such as text string for the text shape).
	ExtraShapeData ShapeData
}

// Marshal ...
func (x *DebugDrawerShape) Marshal(io IO) {
	io.Varuint64(&x.NetworkID)
	OptionalFunc(io, &x.Type, io.Uint8)
	OptionalFunc(io, &x.Location, io.Vec3)
	OptionalFunc(io, &x.Scale, io.Float32)
	OptionalFunc(io, &x.Rotation, io.Vec3)
	OptionalFunc(io, &x.TotalTimeLeft, io.Float32)
	OptionalFunc(io, &x.Colour, io.BEARGB)
	OptionalFunc(io, &x.DimensionID, io.Varint32)
	OptionalFunc(io, &x.AttachedToEntityID, io.Varuint64)
	io.ShapeData(&x.ExtraShapeData)
}
