package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Disconnect may be sent by the server to disconnect the client using an optional message to send as the
// disconnect screen.
type Disconnect struct {
	// Reason is the reason for the disconnection. It seems as if this field has no use other than for
	// telemetry reasons as it does not affect the message that gets displayed on the disconnect screen.
	Reason int32
	// HideDisconnectionScreen specifies if the disconnection screen should be hidden when the client is
	// disconnected, meaning it will be sent directly to the main menu.
	HideDisconnectionScreen bool
	// Message is an optional message to show when disconnected. This message is only written if the
	// HideDisconnectionScreen field is set to true.
	Message string
	// FilteredMessage is always set to empty and the usage is currently unknown.
	FilteredMessage string
}

// ID ...
func (*Disconnect) ID() uint32 {
	return IDDisconnect
}

func (pk *Disconnect) Marshal(io protocol.IO) {
	io.Varint32(&pk.Reason)
	io.Bool(&pk.HideDisconnectionScreen)
	if !pk.HideDisconnectionScreen {
		io.String(&pk.Message)
		io.String(&pk.FilteredMessage)
	}
}
