package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	EnableMultiPlayer = iota
	DisableMultiPlayer
	RefreshJoinCode
)

// MultiPlayerSettings is sent by the client to update multi-player related settings server-side and sent back
// to online players by the server.
// The MultiPlayerSettings packet is a Minecraft: Education Edition packet. It has no functionality for the
// base game.
type MultiPlayerSettings struct {
	// ActionType is the action that should be done when this packet is sent. It is one of the constants that
	// may be found above.
	ActionType int32
}

// ID ...
func (*MultiPlayerSettings) ID() uint32 {
	return IDMultiPlayerSettings
}

// Marshal ...
func (pk *MultiPlayerSettings) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.ActionType)
}

// Unmarshal ...
func (pk *MultiPlayerSettings) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.ActionType)
}
