package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Disconnect may be sent by the server to disconnect the client using an optional message to send as the
// disconnect screen.
type Disconnect struct {
	// HideDisconnectionScreen specifies if the disconnection screen should be hidden when the client is
	// disconnected, meaning it will be sent directly to the main menu.
	HideDisconnectionScreen bool
	// Message is an optional message to show when disconnected. This message is only written if the
	// HideDisconnectionScreen field is set to true.
	Message string
}

// ID ...
func (*Disconnect) ID() uint32 {
	return IDDisconnect
}

func (pk *Disconnect) Marshal(io protocol.IO) {
	io.Bool(&pk.HideDisconnectionScreen)
	if !pk.HideDisconnectionScreen {
		io.String(&pk.Message)
	}
}
