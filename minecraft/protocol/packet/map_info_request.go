package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// MapInfoRequest is sent by the client to request the server to deliver information of a certain map in the
// inventory of the player. The server should respond with a ClientBoundMapItemData packet.
type MapInfoRequest struct {
	// MapID is the unique identifier that represents the map that is requested over network. It remains
	// consistent across sessions.
	MapID int64
	// ClientPixels is a slice of pixels sent from the client to notify the server about the pixels that it isn't aware
	// of.
	ClientPixels []protocol.PixelRequest
}

// ID ...
func (*MapInfoRequest) ID() uint32 {
	return IDMapInfoRequest
}

// Marshal ...
func (pk *MapInfoRequest) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.MapID)
	protocol.SliceUint32Length(w, &pk.ClientPixels)
}

// Unmarshal ...
func (pk *MapInfoRequest) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.MapID)
	protocol.SliceUint32Length(r, &pk.ClientPixels)
}
