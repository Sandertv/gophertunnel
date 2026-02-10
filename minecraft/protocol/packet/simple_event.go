package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SimpleEventCommandsEnabled = iota + 1
	SimpleEventCommandsDisabled
	SimpleEventUnlockWorldTemplateSettings
)

// SimpleEvent is used for enabling or disabling commands and for unlocking world template settings
// (both unlocking UI buttons on client and the actual setting on the server).
// This is fired from the client to the server and a SetCommandsEnabled is sent back when enabling commands.
type SimpleEvent struct {
	// EventType is the type of the event to be called. It is one of the constants that may be found above.
	EventType uint16
}

// ID ...
func (*SimpleEvent) ID() uint32 {
	return IDSimpleEvent
}

func (pk *SimpleEvent) Marshal(io protocol.IO) {
	io.Uint16(&pk.EventType)
}
