package protocol

import (
	"image/color"
)

const (
	MapDecorationTypeMarkerWhite = iota
	MapDecorationTypeMarkerGreen
	MapDecorationTypeMarkerRed
	MapDecorationTypeMarkerBlue
	MapDecorationTypeCrossWhite
	MapDecorationTypeTriangleRed
	MapDecorationTypeSquareWhite
	MapDecorationTypeMarkerSign
	MapDecorationTypeMarkerPink
	MapDecorationTypeMarkerOrange
	MapDecorationTypeMarkerYellow
	MapDecorationTypeMarkerTeal
	MapDecorationTypeTriangleGreen
	MapDecorationTypeSmallSquareWhite
	MapDecorationTypeMansion
	MapDecorationTypeMonument
	MapDecorationTypeNoDraw
	MapDecorationTypeVillageDesert
	MapDecorationTypeVillagePlains
	MapDecorationTypeVillageSavanna
	MapDecorationTypeVillageSnowy
	MapDecorationTypeVillageTaiga
	MapDecorationTypeJungleTemple
	MapDecorationTypeWitchHut
)

const (
	MapObjectTypeEntity = iota
	MapObjectTypeBlock
)

// MapTrackedObject is an object on a map that is 'tracked' by the client, such as an entity or a block. This
// object may move, which is handled client-side.
type MapTrackedObject struct {
	// Type is the type of the tracked object. It is either MapObjectTypeEntity or MapObjectTypeBlock.
	Type int32
	// EntityUniqueID is the optional unique ID of the tracked entity.
	EntityUniqueID Optional[int64]
	// BlockPosition is the optional position of the tracked block.
	BlockPosition Optional[BlockPos]
}

// Marshal encodes/decodes a MapTrackedObject.
func (x *MapTrackedObject) Marshal(r IO) {
	r.Int32(&x.Type)
	OptionalFunc(r, &x.EntityUniqueID, r.ActorUniqueID)
	OptionalFunc(r, &x.BlockPosition, r.BlockPos)
	if x.Type != MapObjectTypeEntity && x.Type != MapObjectTypeBlock {
		r.UnknownEnumOption(x.Type, "map tracked object type")
	}
}

// MapDecoration is a fixed decoration on a map: Its position or other properties do not change automatically
// client-side.
type MapDecoration struct {
	// Type is the type of the map decoration. The type specifies the shape (and sometimes the colour) that
	// the map decoration gets. It is one of the MapDecorationType constants above.
	Type byte
	// Rotation is the rotation of the map decoration. It is byte due to the 16 fixed directions that the
	// map decoration may face.
	Rotation byte
	// X is the offset on the X axis in pixels of the decoration.
	X byte
	// Y is the offset on the Y axis in pixels of the decoration.
	Y byte
	// Label is the name of the map decoration. This name may be of any value.
	Label string
	// Colour is the colour of the map decoration. Some map decoration types have a specific colour set
	// automatically, whereas others may be changed.
	Colour color.RGBA
}

// Marshal encodes/decodes a MapDecoration.
func (x *MapDecoration) Marshal(r IO) {
	r.Uint8(&x.Type)
	r.Uint8(&x.Rotation)
	r.Uint8(&x.X)
	r.Uint8(&x.Y)
	r.String(&x.Label)
	r.BEARGB(&x.Colour)
}

// PixelRequest is the request for the colour of a pixel in a MapInfoRequest packet.
type PixelRequest struct {
	Colour color.RGBA
	Index  uint16
}

// Marshal encodes/decodes a PixelRequest.
func (x *PixelRequest) Marshal(r IO) {
	r.RGBA(&x.Colour)
	r.Uint16(&x.Index)
}
