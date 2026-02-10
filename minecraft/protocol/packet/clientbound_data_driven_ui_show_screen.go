package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundDataDrivenUIShowScreen is sent by the server to show a data-driven UI screen on the client.
type ClientBoundDataDrivenUIShowScreen struct {
	// ScreenID is the identifier of the screen to show.
	ScreenID string
}

// ID ...
func (*ClientBoundDataDrivenUIShowScreen) ID() uint32 {
	return IDClientBoundDataDrivenUIShowScreen
}

func (pk *ClientBoundDataDrivenUIShowScreen) Marshal(io protocol.IO) {
	io.String(&pk.ScreenID)
}
