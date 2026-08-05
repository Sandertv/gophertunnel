package packet

import (
	"image/color"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundMapItemData is sent by the server to the client to update the data of a map shown to the client.
// It is sent with a combination of flags that specify what data is updated.
// The ClientBoundMapItemData packet may be used to update specific parts of the map only. It is not required
// to send the entire map each time when updating one part.
type ClientBoundMapItemData struct {
	// MapID is the unique identifier that represents the map that is updated over network. It remains
	// consistent across sessions.
	MapID int64
	// Dimension is the dimension of the map that should be updated, for example the overworld (0), the nether
	// (1) or the end (2).
	Dimension byte
	// LockedMap specifies if the map that was updated was a locked map, which may be done using a cartography
	// table.
	LockedMap bool
	// Origin is the center position of the map being updated.
	Origin protocol.BlockPos
	// Scale is the scale of the map as it is shown in-game.
	Scale protocol.Optional[byte]
	// MapsIncludedIn holds an array of map IDs that the map updated is included in. This has to do with the
	// scale of the map: Each map holds its own map ID and all map IDs of maps that include this map and have
	// a bigger scale. This means that a scale 0 map will have 5 map IDs in this slice, whereas a scale 4 map
	// will have only 1 (its own).
	// The actual use of this field remains unknown.
	MapsIncludedIn protocol.Optional[[]int64]
	// TrackedObjects is a list of tracked objects on the map, which may either be entities or blocks. The
	// client makes sure these tracked objects are actually tracked. (position updated etc.)
	TrackedObjects protocol.Optional[[]protocol.MapTrackedObject]
	// Decorations is a list of fixed decorations located on the map. The decorations will not change
	// client-side, unless the server updates them.
	Decorations protocol.Optional[[]protocol.MapDecoration]

	// Height is the height of the texture area that was updated. The height may be a subset of the total
	// height of the map.
	Height protocol.Optional[int32]
	// Width is the width of the texture area that was updated. The width may be a subset of the total width
	// of the map.
	Width protocol.Optional[int32]
	// XOffset is the X offset in pixels at which the updated texture area starts. From this X, the updated
	// texture will extend exactly Width pixels to the right.
	XOffset protocol.Optional[int32]
	// YOffset is the Y offset in pixels at which the updated texture area starts. From this Y, the updated
	// texture will extend exactly Height pixels up.
	YOffset protocol.Optional[int32]
	// Pixels is a list of pixel colours for the new texture of the map. It is indexed as Pixels[y*height + x].
	Pixels protocol.Optional[[]color.RGBA]
}

// ID ...
func (*ClientBoundMapItemData) ID() uint32 {
	return IDClientBoundMapItemData
}

func (pk *ClientBoundMapItemData) Marshal(io protocol.IO) {
	io.Varint64(&pk.MapID)
	io.Uint8(&pk.Dimension)
	io.Bool(&pk.LockedMap)
	io.BlockPos(&pk.Origin)

	protocol.OptionalFunc(io, &pk.MapsIncludedIn, func(x *[]int64) {
		protocol.FuncSlice(io, x, io.Varint64)
	})
	protocol.OptionalFunc(io, &pk.Scale, io.Uint8)
	protocol.OptionalFunc(io, &pk.TrackedObjects, func(x *[]protocol.MapTrackedObject) {
		protocol.Slice(io, x)
	})
	protocol.OptionalFunc(io, &pk.Decorations, func(x *[]protocol.MapDecoration) {
		protocol.Slice(io, x)
	})
	protocol.OptionalFunc(io, &pk.Width, io.Varint32)
	protocol.OptionalFunc(io, &pk.Height, io.Varint32)
	protocol.OptionalFunc(io, &pk.XOffset, io.Varint32)
	protocol.OptionalFunc(io, &pk.YOffset, io.Varint32)
	protocol.OptionalFunc(io, &pk.Pixels, func(x *[]color.RGBA) {
		protocol.FuncSlice(io, x, io.RGBA)
	})
}
