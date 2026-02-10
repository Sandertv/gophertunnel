package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundDataDrivenUICloseAllScreens is sent by the server to close all data-driven UI screens on the client.
type ClientBoundDataDrivenUICloseAllScreens struct{}

// ID ...
func (*ClientBoundDataDrivenUICloseAllScreens) ID() uint32 {
	return IDClientBoundDataDrivenUICloseAllScreens
}

func (pk *ClientBoundDataDrivenUICloseAllScreens) Marshal(protocol.IO) {}
