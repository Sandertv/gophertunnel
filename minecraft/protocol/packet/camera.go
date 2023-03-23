package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Camera is sent by the server to use an Education Edition camera on a player. It produces an image
// client-side.
type Camera struct {
	// CameraEntityUniqueID is the unique ID of the camera entity from which the picture was taken.
	CameraEntityUniqueID int64
	// TargetPlayerUniqueID is the unique ID of the target player. The unique ID is a value that remains
	// consistent across different sessions of the same world, but most servers simply fill the runtime ID of
	// the player out for this field.
	TargetPlayerUniqueID int64
}

// ID ...
func (*Camera) ID() uint32 {
	return IDCamera
}

func (pk *Camera) Marshal(io protocol.IO) {
	io.Varint64(&pk.CameraEntityUniqueID)
	io.Varint64(&pk.TargetPlayerUniqueID)
}
