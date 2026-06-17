package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	DataDrivenScreenCloseReasonProgrammaticClose    = "programmaticclose"
	DataDrivenScreenCloseReasonProgrammaticCloseAll = "programmaticcloseall"
	DataDrivenScreenCloseReasonClientCanceled       = "clientcanceled"
	DataDrivenScreenCloseReasonUserBusy             = "userbusy"
	DataDrivenScreenCloseReasonInvalidForm          = "invalidform"
)

// ServerBoundDataDrivenScreenClosed is sent by the client when a data-driven UI screen is closed.
type ServerBoundDataDrivenScreenClosed struct {
	// FormID is the optional unique instance ID of the form that was closed.
	FormID protocol.Optional[uint32]
	// CloseReason is the reason the screen was closed. It is one of the DataDrivenScreenCloseReason constants.
	CloseReason string
}

// ID ...
func (*ServerBoundDataDrivenScreenClosed) ID() uint32 {
	return IDServerBoundDataDrivenScreenClosed
}

func (pk *ServerBoundDataDrivenScreenClosed) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.FormID, io.Uint32)
	io.String(&pk.CloseReason)
}
