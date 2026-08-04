package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ScoreboardActionModify = iota
	ScoreboardActionRemove
)

// SetScore is sent by the server to send the contents of a scoreboard to the player. It may be used to either
// add, remove or edit entries on the scoreboard.
type SetScore struct {
	// ActionType is the type of the action to execute upon the scoreboard with the entries that the packet
	// has. If ActionType is ScoreboardActionModify, all entries will be added to the scoreboard if not yet
	// present, or modified if already present. If set to ScoreboardActionRemove, all scoreboard entries set
	// will be removed from the scoreboard.
	ActionType byte
	// Entries is a list of all entries that the client should operate on. When modifying, it will add or
	// modify all entries, whereas when removing, it will remove all entries.
	Entries []protocol.ScoreboardEntry
}

// ID ...
func (*SetScore) ID() uint32 {
	return IDSetScore
}

func (pk *SetScore) Marshal(io protocol.IO) {
	if pk.ActionType == ScoreboardActionRemove {
		for i := range pk.Entries {
			pk.Entries[i].IdentityType = protocol.ScoreboardIdentityRemove
		}
	}
	protocol.Slice(io, &pk.Entries)
	if _, reading := io.(*protocol.Reader); reading {
		pk.ActionType = ScoreboardActionModify
		if len(pk.Entries) != 0 {
			pk.ActionType = ScoreboardActionRemove
			for _, entry := range pk.Entries {
				if entry.IdentityType != protocol.ScoreboardIdentityRemove {
					pk.ActionType = ScoreboardActionModify
					break
				}
			}
		}
	}
}
