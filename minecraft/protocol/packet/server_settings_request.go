package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerSettingsRequest is sent by the client to request the settings specific to the server. These settings
// are shown in a separate tab client-side, and have the same structure as a custom form.
type ServerSettingsRequest struct {
	// ServerSettingsRequest has no fields.
}

// ID ...
func (*ServerSettingsRequest) ID() uint32 {
	return IDServerSettingsRequest
}

func (*ServerSettingsRequest) Marshal(protocol.IO) {}
