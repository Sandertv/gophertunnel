package packet

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerBoundPackSettingChange is sent by the client to the server when it changes Pack Settings (pack UI).
type ServerBoundPackSettingChange struct {
	// PackID ...
	PackID uuid.UUID
	// PackSetting ...
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
