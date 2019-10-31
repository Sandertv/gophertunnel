package packet

import (
	"bytes"
	"encoding/binary"
)

const (
	EnableMultiPlayer = iota
	DisableMultiPlayer
	RefreshJoinCode
)

// MultiPlayerSettings is sent by the server to update multi-player related settings. Usually these settings
// are also sent in the StartGame packet.
// The MultiPlayerSettings packet is a Minecraft: Education Edition packet. It has no functionality for the
// base game.
type MultiPlayerSettings struct {
	// ActionType is the action that should be done when this packet is sent. It is one of the constants that
	// may be found above.
	ActionType byte
}

// ID ...
func (*MultiPlayerSettings) ID() uint32 {
	return IDMultiPlayerSettings
}

// Marshal ...
func (pk *MultiPlayerSettings) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
}

// Unmarshal ...
func (pk *MultiPlayerSettings) Unmarshal(buf *bytes.Buffer) error {
	return binary.Write(buf, binary.LittleEndian, &pk.ActionType)
}
