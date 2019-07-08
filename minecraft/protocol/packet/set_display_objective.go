package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ScoreboardSlotList      = "list"
	ScoreboardSlotSidebar   = "sidebar"
	ScoreboardSlotBelowName = "belowname"
)

const (
	ScoreboardSortOrderAscending = iota
	ScoreboardSortOrderDescending
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
func (pk *SetDisplayObjective) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.DisplaySlot)
	_ = protocol.WriteString(buf, pk.ObjectiveName)
	_ = protocol.WriteString(buf, pk.DisplayName)
	_ = protocol.WriteString(buf, pk.CriteriaName)
	_ = protocol.WriteVarint32(buf, pk.SortOrder)
}

// Unmarshal ...
func (pk *SetDisplayObjective) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.DisplaySlot),
		protocol.String(buf, &pk.ObjectiveName),
		protocol.String(buf, &pk.DisplayName),
		protocol.String(buf, &pk.CriteriaName),
		protocol.Varint32(buf, &pk.SortOrder),
	)
}
