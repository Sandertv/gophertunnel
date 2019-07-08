package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	ScoreboardIdentityPlayer = iota + 1
	ScoreboardIdentityEntity
	ScoreboardIdentityFakePlayer
)

// ScoreboardEntry represents a single entry that may be found on a scoreboard. These entries represent a
// line on the scoreboard each.
type ScoreboardEntry struct {
	// EntryID is a unique identifier of this entry. Each entry that represents a different value should get
	// its own entry ID. When modifying a scoreboard, entries that represent the same line should have the
	// same entry ID.
	EntryID int64
	// ObjectiveName is the name of the objective that this scoreboard entry is for. It must be identical to
	// the one set in the SetDisplayObjective packet previously sent.
	ObjectiveName string
	// Score is the score that the entry represents. Scoreboard entries are ordered using this score, so in
	// order to get the scoreboard to be ordered as expected when trying to write non-score related text on
	// a scoreboard, this score should be incremented for each entry.
	Score int32
	// IdentityType is the identity type of the scoreboard entry. The entry may represent an entity, player or
	// a fake player, as the constants above indicate.
	// In order to write plain text to the scoreboard, ScoreboardIdentityFakePlayer should always be used, in
	// combination with the DisplayName field. A different identity type will use the name of the entity.
	IdentityType byte
	// EntityUniqueID is the unique ID of either the player or the entity represented by the scoreboard entry.
	// This field is only used if IdentityType is either ScoreboardIdentityEntity or ScoreboardIdentityPlayer.
	EntityUniqueID int64
	// DisplayName is the custom name of the scoreboard entry. This field is only used if IdentityType is
	// ScoreboardIdentityFakePlayer. If this identity type is not used, the name of the entity/player will be
	// shown instead.
	DisplayName string
}

// WriteScoreEntry writes a ScoreboardEntry x to Buffer dst. If modify is set to true, the display information
// of the entry is written. If not, it is ignored, as expected when the SetScore packet is sent to modify
// entries.
func WriteScoreEntry(dst *bytes.Buffer, x ScoreboardEntry, modify bool) error {
	if err := chainErr(
		WriteVarint64(dst, x.EntryID),
		WriteString(dst, x.ObjectiveName),
		binary.Write(dst, binary.LittleEndian, x.Score),
	); err != nil {
		return err
	}
	if modify {
		if err := binary.Write(dst, binary.LittleEndian, x.IdentityType); err != nil {
			return err
		}
		switch x.IdentityType {
		case ScoreboardIdentityEntity, ScoreboardIdentityPlayer:
			return WriteVarint64(dst, x.EntityUniqueID)
		case ScoreboardIdentityFakePlayer:
			return WriteString(dst, x.DisplayName)
		default:
			panic(fmt.Sprintf("invalid scoreboardy entry identity type %v", x.IdentityType))
		}
	}
	return nil
}

// ScoreEntry reads a ScoreboardEntry x from Buffer src. It reads the display information if modify is true,
// as expected when the SetScore packet is sent to modify entries.
func ScoreEntry(src *bytes.Buffer, x *ScoreboardEntry, modify bool) error {
	if err := chainErr(
		Varint64(src, &x.EntryID),
		String(src, &x.ObjectiveName),
		binary.Read(src, binary.LittleEndian, &x.Score),
	); err != nil {
		return err
	}
	if modify {
		if err := binary.Read(src, binary.LittleEndian, &x.IdentityType); err != nil {
			return err
		}
		switch x.IdentityType {
		case ScoreboardIdentityEntity, ScoreboardIdentityPlayer:
			return Varint64(src, &x.EntityUniqueID)
		case ScoreboardIdentityFakePlayer:
			return String(src, &x.DisplayName)
		default:
			return fmt.Errorf("unknown scoreboard identity type %v", x.IdentityType)
		}
	}
	return nil
}
