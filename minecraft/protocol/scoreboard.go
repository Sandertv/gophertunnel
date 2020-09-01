package protocol

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

// ScoreboardIdentityEntry holds an entry to either associate an identity with one of the entries in a
// scoreboard, or to remove associations.
type ScoreboardIdentityEntry struct {
	// EntryID is the unique identifier of the entry that the identity should be associated with, or that
	// associations should be cleared from.
	EntryID int64
	// EntityUniqueID is the unique ID that the entry should be associated with. It is empty if the
	// SetScoreboardIdentity packet is sent to remove associations with identities.
	EntityUniqueID int64
}

// ScoreEntry reads/writes a ScoreboardEntry x using IO r. It reads the display information if modify is true,
// as expected when the SetScore packet is sent to modify entries.
func ScoreEntry(r IO, x *ScoreboardEntry, modify bool) {
	r.Varint64(&x.EntryID)
	r.String(&x.ObjectiveName)
	r.Int32(&x.Score)

	if modify {
		r.Uint8(&x.IdentityType)
		switch x.IdentityType {
		case ScoreboardIdentityEntity, ScoreboardIdentityPlayer:
			r.Varint64(&x.EntityUniqueID)
		case ScoreboardIdentityFakePlayer:
			r.String(&x.DisplayName)
		default:
			r.UnknownEnumOption(x.IdentityType, "scoreboard entry identity type")
		}
	}
}
