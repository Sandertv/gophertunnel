package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdatePlayerGameType is sent by the server to change the game mode of a player. It is functionally
// identical to the SetPlayerGameType packet.
type UpdatePlayerGameType struct {
	// GameType is the new game type of the player. It is one of the constants that can be found in
	// set_player_game_type.go. Some of these game types require additional flags to be set in an
	// AdventureSettings packet for the game mode to obtain its full functionality.
	GameType int32
	// PlayerUniqueID is the entity unique ID of the player that should have its game mode updated. If this
	// packet is sent to other clients with the player unique ID of another player, nothing happens.
	PlayerUniqueID int64
}

// ID ...
func (*UpdatePlayerGameType) ID() uint32 {
	return IDUpdatePlayerGameType
}

// Marshal ...
func (pk *UpdatePlayerGameType) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.GameType)
	_ = protocol.WriteVarint64(buf, pk.PlayerUniqueID)
}

// Unmarshal ...
func (pk *UpdatePlayerGameType) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.Varint32(buf, &pk.GameType),
		protocol.Varint64(buf, &pk.PlayerUniqueID),
	)
}
