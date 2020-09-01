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
	l := uint32(len(pk.Entries))
	w.Uint8(&pk.ActionType)
	w.Varuint32(&l)
	for _, entry := range pk.Entries {
		switch pk.ActionType {
		case PlayerListActionAdd:
			protocol.WritePlayerAddEntry(w, &entry)
		case PlayerListActionRemove:
			w.UUID(&entry.UUID)
		default:
			w.UnknownEnumOption(pk.ActionType, "player list action type")
		}
	}
	if pk.ActionType == PlayerListActionAdd {
		for _, entry := range pk.Entries {
			w.Bool(&entry.Skin.Trusted)
		}
	}
}

// Unmarshal ...
func (pk *PlayerList) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Uint8(&pk.ActionType)
	r.Varuint32(&count)

	pk.Entries = make([]protocol.PlayerListEntry, count)
	for i := uint32(0); i < count; i++ {
		switch pk.ActionType {
		case PlayerListActionAdd:
			protocol.PlayerAddEntry(r, &pk.Entries[i])
		case PlayerListActionRemove:
			r.UUID(&pk.Entries[i].UUID)
		default:
			r.UnknownEnumOption(pk.ActionType, "player list action type")
		}
	}
	if pk.ActionType == PlayerListActionAdd {
		for i := uint32(0); i < count; i++ {
			r.Bool(&pk.Entries[i].Skin.Trusted)
		}
	}
}
