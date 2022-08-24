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
	case ScoreboardActionRemove:
		protocol.FuncSliceIO(w, &pk.Entries, protocol.ScoreRemoveEntry)
	case ScoreboardActionModify:
		protocol.FuncSliceIO(w, &pk.Entries, protocol.ScoreModifyEntry)
	default:
		w.UnknownEnumOption(pk.ActionType, "set score action type")
	}
}

// Unmarshal ...
func (pk *SetScore) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case ScoreboardActionRemove:
		protocol.FuncSliceIO(r, &pk.Entries, protocol.ScoreRemoveEntry)
	case ScoreboardActionModify:
		protocol.FuncSliceIO(r, &pk.Entries, protocol.ScoreModifyEntry)
	default:
		r.UnknownEnumOption(pk.ActionType, "set score action type")
	}
}
