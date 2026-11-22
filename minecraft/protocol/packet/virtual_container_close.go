package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	// VirtualContainerCloseKindRequest is sent by the server to ask the client to close a
	// virtual container that is currently displayed.
	VirtualContainerCloseKindRequest uint8 = iota
	// VirtualContainerCloseKindClosed is sent by the client when the container could not be
	// opened, when the client closed it, or in response to a request from the server.
	VirtualContainerCloseKindClosed
	// VirtualContainerCloseKindConfirmed is sent by the server once the player's pending
	// inventory requests have been processed.
	VirtualContainerCloseKindConfirmed
)

const (
	VirtualContainerCauseUnknown uint8 = iota
	VirtualContainerCauseUserClosed
	VirtualContainerCauseServerRequested
	VirtualContainerCauseUserBusy
	VirtualContainerCauseClientNotReady
	VirtualContainerCauseInvalidContents
	VirtualContainerCauseUICreationFailed
	VirtualContainerCauseUICanceled
)

// VirtualContainerClose is sent by both server and client to close a virtual container.
type VirtualContainerClose struct {
	// SessionID is the ID of the virtual container being closed.
	SessionID uint64
	// Kind is the kind of the closure. It is one of the constants above.
	Kind uint8
	// Cause is the reason the virtual container was closed or could not be opened.
	// It is one of the constants above.
	Cause uint8
}

// ID ...
func (*VirtualContainerClose) ID() uint32 {
	return IDVirtualContainerClose
}

func (pk *VirtualContainerClose) Marshal(io protocol.IO) {
	io.Varuint64(&pk.SessionID)
	io.Uint8(&pk.Kind)
	io.Uint8(&pk.Cause)
}
