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

// Marshal ...
func (pk *PlayerList) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *PlayerList) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *PlayerList) marshal(r protocol.IO) {
	r.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case PlayerListActionAdd:
		protocol.Slice(r, &pk.Entries)
	case PlayerListActionRemove:
		protocol.FuncIOSlice(r, &pk.Entries, protocol.PlayerListRemoveEntry)
	default:
		r.UnknownEnumOption(pk.ActionType, "player list action type")
	}
	if pk.ActionType == PlayerListActionAdd {
		for i := 0; i < len(pk.Entries); i++ {
			r.Bool(&pk.Entries[i].Skin.Trusted)
		}
	}
}
