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
}

// ID ...
func (*MapInfoRequest) ID() uint32 {
	return IDMapInfoRequest
}

// Marshal ...
func (pk *MapInfoRequest) Marshal(w *protocol.Writer) {
	w.Varint64(&pk.MapID)
}

// Unmarshal ...
func (pk *MapInfoRequest) Unmarshal(r *protocol.Reader) {
	r.Varint64(&pk.MapID)
}
