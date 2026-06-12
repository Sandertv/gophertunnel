package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	DataDrivenScreenCloseReasonProgrammaticClose = iota
	DataDrivenScreenCloseReasonProgrammaticCloseAll
	DataDrivenScreenCloseReasonClientCanceled
	DataDrivenScreenCloseReasonUserBusy
	DataDrivenScreenCloseReasonInvalidForm
)

// ServerBoundDataDrivenScreenClosed is sent by the client when a data-driven UI screen is closed.
type ServerBoundDataDrivenScreenClosed struct {
	// FormID is the optional unique instance ID of the form that was closed.
	FormID protocol.Optional[uint32]
	// CloseReason is the reason the screen was closed. It is one of the DataDrivenScreenCloseReason constants.
	CloseReason uint8
}

// ID ...
func (*ServerBoundDataDrivenScreenClosed) ID() uint32 {
	return IDServerBoundDataDrivenScreenClosed
}

func (pk *ServerBoundDataDrivenScreenClosed) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.FormID, io.Uint32)
	io.Uint8(&pk.CloseReason)
}
