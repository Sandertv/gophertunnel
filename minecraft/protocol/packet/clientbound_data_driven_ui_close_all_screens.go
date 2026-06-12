package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundDataDrivenUICloseScreen is sent by the server to close a data-driven UI screen on the client.
// If FormID is not set, all data-driven UI screens are closed.
type ClientBoundDataDrivenUICloseScreen struct {
	// FormID is the optional unique instance ID of the form to close. If not set, all forms are closed.
	FormID protocol.Optional[uint32]
}

// ID ...
func (*ClientBoundDataDrivenUICloseScreen) ID() uint32 {
	return IDClientBoundDataDrivenUICloseScreen
}

func (pk *ClientBoundDataDrivenUICloseScreen) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.FormID, io.Uint32)
}
