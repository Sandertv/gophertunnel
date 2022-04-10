package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	GameTypeSurvival = iota
	GameTypeCreative
	GameTypeAdventure
	GameTypeSurvivalSpectator
	GameTypeCreativeSpectator
	GameTypeDefault
	GameTypeSpectator
)

// SetPlayerGameType is sent by the server to update the game type, which is otherwise known as the game mode,
// of a player.
type SetPlayerGameType struct {
	// GameType is the new game type of the player. It is one of the constants that can be found above. Some
	// of these game types require additional flags to be set in an AdventureSettings packet for the game mode
	// to obtain its full functionality.
	GameType int32
}

// ID ...
func (*SetPlayerGameType) ID() uint32 {
	return IDSetPlayerGameType
}

// Marshal ...
func (pk *SetPlayerGameType) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.GameType)
}

// Unmarshal ...
func (pk *SetPlayerGameType) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.GameType)
}
