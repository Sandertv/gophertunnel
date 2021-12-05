package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ScoreboardSortOrderAscending = iota
	ScoreboardSortOrderDescending
)

//noinspection SpellCheckingInspection
const (
	ScoreboardSlotList      = "list"
	ScoreboardSlotSidebar   = "sidebar"
	ScoreboardSlotBelowName = "belowname"
)

// SetDisplayObjective is sent by the server to display an object as a scoreboard to the player. Once sent,
// it should be followed up by a SetScore packet to set the lines of the packet.
type SetDisplayObjective struct {
	// DisplaySlot is the slot in which the scoreboard should be displayed. Available options can be found in
	// the constants above.
	DisplaySlot string
	// ObjectiveName is the name of the objective that the scoreboard displays. Filling out a random unique
	// value for this field works: It is not displayed in the scoreboard.
	ObjectiveName string
	// DisplayName is the name, or title, that is displayed at the top of the scoreboard.
	DisplayName string
	// CriteriaName is the name of the criteria that need to be fulfilled in order for the score to be
	// increased. This can be any kind of string and does not show up client-side.
	CriteriaName string
	// SortOrder is the order in which entries on the scoreboard should be sorted. It is one of the constants
	// that may be found above.
	SortOrder int32
}

// ID ...
func (*SetDisplayObjective) ID() uint32 {
	return IDSetDisplayObjective
}

// Marshal ...
func (pk *SetDisplayObjective) Marshal(w *protocol.Writer) {
	w.String(&pk.DisplaySlot)
	w.String(&pk.ObjectiveName)
	w.String(&pk.DisplayName)
	w.String(&pk.CriteriaName)
	w.Varint32(&pk.SortOrder)
}

// Unmarshal ...
func (pk *SetDisplayObjective) Unmarshal(r *protocol.Reader) {
	r.String(&pk.DisplaySlot)
	r.String(&pk.ObjectiveName)
	r.String(&pk.DisplayName)
	r.String(&pk.CriteriaName)
	r.Varint32(&pk.SortOrder)
}
