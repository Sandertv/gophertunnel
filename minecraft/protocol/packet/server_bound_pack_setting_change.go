package packet

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerBoundPackSettingChange is sent by the client to the server when it changes a setting
// for a specific pack in the pack settings UI.
type ServerBoundPackSettingChange struct {
	// PackID is the UUID of the pack.
	PackID uuid.UUID
	// PackSetting is the new setting value applied to the pack.
	PackSetting protocol.PackSetting
}

// ID ...
func (*ServerBoundPackSettingChange) ID() uint32 {
	return IDServerBoundPackSettingChange
}

func (pk *ServerBoundPackSettingChange) Marshal(io protocol.IO) {
	io.UUID(&pk.PackID)
	io.PackSetting(&pk.PackSetting)
}
