package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundDataDrivenUIShowScreen is sent by the server to show a data-driven UI screen on the client.
type ClientBoundDataDrivenUIShowScreen struct {
	// ScreenID is the identifier of the screen to show.
	ScreenID string
	// FormID is a unique instance ID for the form, used for scripting to identify specific screen instances.
	FormID uint32
	// DataInstanceID is an optional data ID associated with the screen.
	DataInstanceID protocol.Optional[uint32]
}

// ID ...
func (*ClientBoundDataDrivenUIShowScreen) ID() uint32 {
	return IDClientBoundDataDrivenUIShowScreen
}

func (pk *ClientBoundDataDrivenUIShowScreen) Marshal(io protocol.IO) {
	io.String(&pk.ScreenID)
	io.Uint32(&pk.FormID)
	protocol.OptionalFunc(io, &pk.DataInstanceID, io.Uint32)
}
