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

// Marshal ...
func (pk *SetScore) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case ScoreboardActionRemove, ScoreboardActionModify:
		l := uint32(len(pk.Entries))
		w.Varuint32(&l)
		for _, entry := range pk.Entries {
			protocol.ScoreEntry(w, &entry, pk.ActionType == ScoreboardActionModify)
		}
	default:
		w.UnknownEnumOption(pk.ActionType, "set score action type")
	}
}

// Unmarshal ...
func (pk *SetScore) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Uint8(&pk.ActionType)
	r.Varuint32(&count)

	if pk.ActionType != ScoreboardActionRemove && pk.ActionType != ScoreboardActionModify {
		r.UnknownEnumOption(pk.ActionType, "set score action type")
	}
	pk.Entries = make([]protocol.ScoreboardEntry, count)
	for i := uint32(0); i < count; i++ {
		protocol.ScoreEntry(r, &pk.Entries[i], pk.ActionType == ScoreboardActionModify)
	}
}
