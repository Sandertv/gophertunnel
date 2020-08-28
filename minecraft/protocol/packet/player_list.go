package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
func (pk *PlayerList) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Entries)))
	for _, entry := range pk.Entries {
		switch pk.ActionType {
		case PlayerListActionAdd:
			_ = protocol.WritePlayerAddEntry(buf, entry)
		case PlayerListActionRemove:
			_ = protocol.WritePlayerRemoveEntry(buf, entry)
		default:
			panic(fmt.Sprintf("invalid player list action type %v", pk.ActionType))
		}
	}
	if pk.ActionType == PlayerListActionAdd {
		for _, entry := range pk.Entries {
			_ = binary.Write(buf, binary.LittleEndian, entry.Skin.Trusted)
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
			protocol.PlayerRemoveEntry(r, &pk.Entries[i])
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
