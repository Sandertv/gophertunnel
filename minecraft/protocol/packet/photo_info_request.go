package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// PhotoInfoRequest is sent by the client to request photo information from the server.
type PhotoInfoRequest struct {
	// PhotoID is the ID of the photo.
	PhotoID int64
}

// ID ...
func (*PhotoInfoRequest) ID() uint32 {
	return IDPhotoInfoRequest
}

// Marshal ...
func (pk *PhotoInfoRequest) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *PhotoInfoRequest) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *PhotoInfoRequest) marshal(r protocol.IO) {
	r.Varint64(&pk.PhotoID)
}
