package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"image/color"
)

// MapInfoRequest is sent by the client to request the server to deliver information of a certain map in the
// inventory of the player. The server should respond with a ClientBoundMapItemData packet.
type MapInfoRequest struct {
	// MapID is the unique identifier that represents the map that is requested over network. It remains
	// consistent across sessions.
	MapID int64
	// ClientPixels is a map of pixels sent from the client to notify the server about the pixels that it isn't aware of.
	ClientPixels map[uint16]color.RGBA
}

// ID ...
func (*MapInfoRequest) ID() uint32 {
	return IDMapInfoRequest
}

// Marshal ...
func (pk *MapInfoRequest) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.MapID)

	clientPixelsLen := uint32(len(pk.ClientPixels))
	w.Uint32(&clientPixelsLen)
	for index, pixel := range pk.ClientPixels {
		w.Uint16(&index)
		w.VarRGBA(&pixel)
	}
}

// Unmarshal ...
func (pk *MapInfoRequest) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.MapID)

	var clientPixelsLen uint32
	r.Uint32(&clientPixelsLen)

	pk.ClientPixels = make(map[uint16]color.RGBA, clientPixelsLen)
	for i := uint32(0); i < clientPixelsLen; i++ {
		var index uint16
		r.Uint16(&index)

		var pixel color.RGBA
		r.VarRGBA(&pixel)

		pk.ClientPixels[index] = pixel
	}
}
