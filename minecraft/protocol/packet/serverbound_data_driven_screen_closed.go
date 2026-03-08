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

// ServerboundDataDrivenScreenClosed is sent by the client when a data-driven UI screen is closed.
type ServerboundDataDrivenScreenClosed struct {
	// FormID is the optional unique instance ID of the form that was closed.
	FormID protocol.Optional[uint32]
	// CloseReason is the reason the screen was closed. It is one of the DataDrivenScreenCloseReason constants.
	CloseReason uint8
}

// ID ...
func (*ServerboundDataDrivenScreenClosed) ID() uint32 {
	return IDServerboundDataDrivenScreenClosed
}

func (pk *ServerboundDataDrivenScreenClosed) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.FormID, io.Uint32)
	io.Uint8(&pk.CloseReason)
}
