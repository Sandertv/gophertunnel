package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerBoundDataDrivenScreenClosed is sent by the client when a data-driven UI screen is closed.
type ServerBoundDataDrivenScreenClosed struct {
	// FormID is the unique instance ID of the form that was closed.
	FormID uint32
	// CloseReason is the reason the screen was closed.
	CloseReason string
}

// ID ...
func (*ServerBoundDataDrivenScreenClosed) ID() uint32 {
	return IDServerBoundDataDrivenScreenClosed
}

func (pk *ServerBoundDataDrivenScreenClosed) Marshal(io protocol.IO) {
	io.Uint32(&pk.FormID)
	io.String(&pk.CloseReason)
}
