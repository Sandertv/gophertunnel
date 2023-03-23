package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerListActionAdd = iota
	PlayerListActionRemove
)

// PlayerList is sent by the server to update the client-side player list in the in-game menu screen. It shows
// the icon of each player if the correct XUID is written in the packet.
// Sending the PlayerList packet is obligatory when sending an AddPlayer packet. The added player will not
// show up to a client if it has not been added to the player list, because several properties of the player
// are obtained from the player list, such as the skin.
type PlayerList struct {
	// ActionType is the action to execute upon the player list. The entries that follow specify which entries
	// are added or removed from the player list.
	ActionType byte
	// Entries is a list of all player list entries that should be added/removed from the player list,
	// depending on the ActionType set.
	Entries []protocol.PlayerListEntry
}

// ID ...
func (*PlayerList) ID() uint32 {
	return IDPlayerList
}

func (pk *PlayerList) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case PlayerListActionAdd:
		protocol.Slice(io, &pk.Entries)
	case PlayerListActionRemove:
		protocol.FuncIOSlice(io, &pk.Entries, protocol.PlayerListRemoveEntry)
	default:
		io.UnknownEnumOption(pk.ActionType, "player list action type")
	}
	if pk.ActionType == PlayerListActionAdd {
		for i := 0; i < len(pk.Entries); i++ {
			io.Bool(&pk.Entries[i].Skin.Trusted)
		}
	}
}
