package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
func (pk *SetScore) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
	if pk.ActionType != ScoreboardActionRemove && pk.ActionType != ScoreboardActionModify {
		panic(fmt.Sprintf("invalid scoreboard action type %v", pk.ActionType))
	}
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Entries)))
	for _, entry := range pk.Entries {
		_ = protocol.WriteScoreEntry(buf, entry, pk.ActionType == ScoreboardActionModify)
	}
}

// Unmarshal ...
func (pk *SetScore) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.ActionType),
		protocol.Varuint32(buf, &count),
	); err != nil {
		return err
	}
	if pk.ActionType != ScoreboardActionRemove && pk.ActionType != ScoreboardActionModify {
		return fmt.Errorf("unknown scoreboard action type %v", pk.ActionType)
	}
	pk.Entries = make([]protocol.ScoreboardEntry, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.ScoreEntry(buf, &pk.Entries[i], pk.ActionType == ScoreboardActionModify); err != nil {
			return err
		}
	}
	return nil
}
