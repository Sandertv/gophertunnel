package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"image/color"
	"math"
)

const (
	MapUpdateFlagTexture = 1 << (iota + 1)
	MapUpdateFlagDecoration
	MapUpdateFlagInitialisation
)

// ClientBoundMapItemData is sent by the server to the client to update the data of a map shown to the client.
// It is sent with a combination of flags that specify what data is updated.
// The ClientBoundMapItemData packet may be used to update specific parts of the map only. It is not required
// to send the entire map each time when updating one part.
type ClientBoundMapItemData struct {
	// MapID is the unique identifier that represents the map that is updated over network. It remains
	// consistent across sessions.
	MapID int64
	// UpdateFlags is a combination of flags found above that indicate what parts of the map should be updated
	// client-side.
	UpdateFlags uint32
	// Dimension is the dimension of the map that should be updated, for example the overworld (0), the nether
	// (1) or the end (2).
	Dimension byte
	// LockedMap specifies if the map that was updated was a locked map, which may be done using a cartography
	// table.
	LockedMap bool
	// Scale is the scale of the map as it is shown in-game. It is written when any of the MapUpdateFlags are
	// set to the UpdateFlags field.
	Scale byte

	// The following fields apply only for the MapUpdateFlagInitialisation.

	// MapsIncludedIn holds an array of map IDs that the map updated is included in. This has to do with the
	// scale of the map: Each map holds its own map ID and all map IDs of maps that include this map and have
	// a bigger scale. This means that a scale 0 map will have 5 map IDs in this slice, whereas a scale 4 map
	// will have only 1 (its own).
	// The actual use of this field remains unknown.
	MapsIncludedIn []int64

	// The following fields apply only for the MapUpdateFlagDecoration.

	// TrackedObjects is a list of tracked objects on the map, which may either be entities or blocks. The
	// client makes sure these tracked objects are actually tracked. (position updated etc.)
	TrackedObjects []protocol.MapTrackedObject
	// Decorations is a list of fixed decorations located on the map. The decorations will not change
	// client-side, unless the server updates them.
	Decorations []protocol.MapDecoration

	// The following fields apply only for the MapUpdateFlagTexture update flag.

	// Height is the height of the texture area that was updated. The height may be a subset of the total
	// height of the map.
	Height int32
	// Width is the width of the texture area that was updated. The width may be a subset of the total width
	// of the map.
	Width int32
	// XOffset is the X offset in pixels at which the updated texture area starts. From this X, the updated
	// texture will extend exactly Width pixels to the right.
	XOffset int32
	// YOffset is the Y offset in pixels at which the updated texture area starts. From this Y, the updated
	// texture will extend exactly Height pixels up.
	YOffset int32
	// Pixels is a list of pixel colours for the new texture of the map. It is indexed as Pixels[y][x], with
	// the length of the outer slice having to be exactly Height long and the inner slices exactly Width long.
	Pixels [][]color.RGBA
}

// ID ...
func (*ClientBoundMapItemData) ID() uint32 {
	return IDClientBoundMapItemData
}

// Marshal ...
func (pk *ClientBoundMapItemData) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint64(buf, pk.MapID)
	_ = protocol.WriteVaruint32(buf, pk.UpdateFlags)
	_ = binary.Write(buf, binary.LittleEndian, pk.Dimension)
	_ = binary.Write(buf, binary.LittleEndian, pk.LockedMap)

	if pk.UpdateFlags&MapUpdateFlagInitialisation != 0 {
		_ = protocol.WriteVaruint32(buf, uint32(len(pk.MapsIncludedIn)))
		for _, mapID := range pk.MapsIncludedIn {
			_ = protocol.WriteVarint64(buf, mapID)
		}
	}
	if pk.UpdateFlags&(MapUpdateFlagInitialisation|MapUpdateFlagDecoration|MapUpdateFlagTexture) != 0 {
		_ = binary.Write(buf, binary.LittleEndian, pk.Scale)
	}
	if pk.UpdateFlags&MapUpdateFlagDecoration != 0 {
		_ = protocol.WriteVaruint32(buf, uint32(len(pk.TrackedObjects)))
		for _, obj := range pk.TrackedObjects {
			_ = protocol.WriteMapTrackedObj(buf, obj)
		}
		_ = protocol.WriteVaruint32(buf, uint32(len(pk.Decorations)))
		for _, decoration := range pk.Decorations {
			_ = protocol.WriteMapDeco(buf, decoration)
		}
	}
	if pk.UpdateFlags&MapUpdateFlagTexture != 0 {
		// Some basic validation for the values passed into the packet.
		if pk.Width <= 0 || pk.Height <= 0 {
			panic("invalid map texture update: width and height must be at least 1")
		}
		_ = protocol.WriteVarint32(buf, pk.Width)
		_ = protocol.WriteVarint32(buf, pk.Height)
		_ = protocol.WriteVarint32(buf, pk.XOffset)
		_ = protocol.WriteVarint32(buf, pk.YOffset)

		_ = protocol.WriteVaruint32(buf, uint32(pk.Width*pk.Height))

		if len(pk.Pixels) != int(pk.Height) {
			panic("invalid map texture update: length of outer pixels array must be equal to height")
		}
		for y := int32(0); y < pk.Height; y++ {
			if len(pk.Pixels[y]) != int(pk.Width) {
				panic("invalid map texture update: length of inner pixels array must be equal to width")
			}
			for x := int32(0); x < pk.Width; x++ {
				_ = protocol.WriteVarRGBA(buf, pk.Pixels[y][x])
			}
		}
	}
}

// Unmarshal ...
func (pk *ClientBoundMapItemData) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.MapID)
	r.Varuint32(&pk.UpdateFlags)
	r.Uint8(&pk.Dimension)
	r.Bool(&pk.LockedMap)

	var count uint32
	if pk.UpdateFlags&MapUpdateFlagInitialisation != 0 {
		r.Varuint32(&count)
		pk.MapsIncludedIn = make([]int64, count)
		for i := uint32(0); i < count; i++ {
			r.Varint64(&pk.MapsIncludedIn[i])
		}
	}
	if pk.UpdateFlags&(MapUpdateFlagInitialisation|MapUpdateFlagDecoration|MapUpdateFlagTexture) != 0 {
		r.Uint8(&pk.Scale)
	}
	if pk.UpdateFlags&MapUpdateFlagDecoration != 0 {
		r.Varuint32(&count)
		pk.TrackedObjects = make([]protocol.MapTrackedObject, count)
		for i := uint32(0); i < count; i++ {
			protocol.MapTrackedObj(r, &pk.TrackedObjects[i])
		}
		r.Varuint32(&count)
		pk.Decorations = make([]protocol.MapDecoration, count)
		for i := uint32(0); i < count; i++ {
			protocol.MapDeco(r, &pk.Decorations[i])
		}
	}
	if pk.UpdateFlags&MapUpdateFlagTexture != 0 {
		r.Varint32(&pk.Width)
		r.Varint32(&pk.Height)
		r.Varint32(&pk.XOffset)
		r.Varint32(&pk.YOffset)
		r.Varuint32(&count)

		r.LimitInt32(pk.Width, 0, math.MaxInt16)
		r.LimitInt32(pk.Height, 0, math.MaxInt16)
		r.LimitInt32(pk.Width*pk.Height, int32(count), int32(count))

		pk.Pixels = make([][]color.RGBA, pk.Height)
		for y := int32(0); y < pk.Height; y++ {
			pk.Pixels[y] = make([]color.RGBA, pk.Width)
			for x := int32(0); x < pk.Width; x++ {
				protocol.VarRGBA(r, &pk.Pixels[y][x])
			}
		}
	}
}
