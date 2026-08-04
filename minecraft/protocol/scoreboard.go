package protocol

// The constants below identify what a ScoreboardEntry does. An entry either removes a line from the
// scoreboard or writes one, in which case the identity behind the line decides which of the entry's fields
// are used.
const (
	ScoreboardActionRemove uint32 = iota
	ScoreboardActionChangePlayer
	ScoreboardActionChangeEntity
	ScoreboardActionChangeFakePlayer
)

// The names below identify the action of a ScoreboardEntry on the wire. The numeric action is followed by one
// of them.
const (
	ScoreboardActionNameRemove           = "remove"
	ScoreboardActionNameChangePlayer     = "changeplayer"
	ScoreboardActionNameChangeEntity     = "changeentity"
	ScoreboardActionNameChangeFakePlayer = "changefakeplayer"
)

// scoreboardActionNames maps the actions above to their name on the wire, in action order.
var scoreboardActionNames = [...]string{
	ScoreboardActionNameRemove,
	ScoreboardActionNameChangePlayer,
	ScoreboardActionNameChangeEntity,
	ScoreboardActionNameChangeFakePlayer,
}

// ScoreboardEntry represents a single entry that may be found on a scoreboard. These entries represent a
// line on the scoreboard each.
type ScoreboardEntry struct {
	// Action is what this entry does to the scoreboard. It is one of the constants above and decides which of
	// the fields below are written.
	Action uint32
	// EntryID is a unique identifier of this entry. Each entry that represents a different value should get
	// its own entry ID. When modifying a scoreboard, entries that represent the same line should have the
	// same entry ID.
	EntryID int64
	// ObjectiveName is the name of the objective that this scoreboard entry is for. It must be identical to
	// the one set in the SetDisplayObjective packet previously sent. An empty name is only valid for
	// ScoreboardActionRemove, where it removes the entry from every objective.
	ObjectiveName string
	// Score is the score that the entry represents. Scoreboard entries are ordered using this score, so in
	// order to get the scoreboard to be ordered as expected when trying to write non-score related text on
	// a scoreboard, this score should be incremented for each entry. It is not used for
	// ScoreboardActionRemove.
	Score int32
	// EntityUniqueID is the unique ID of either the player or the entity represented by the scoreboard entry.
	// This field is only used if Action is either ScoreboardActionChangePlayer or
	// ScoreboardActionChangeEntity.
	EntityUniqueID int64
	// DisplayName is the custom name of the scoreboard entry. This field is only used if Action is
	// ScoreboardActionChangeFakePlayer. For the other actions the name of the entity/player is shown instead.
	DisplayName string
}

// Marshal encodes/decodes a ScoreboardEntry.
func (x *ScoreboardEntry) Marshal(r IO) {
	r.Varuint32(&x.Action)
	if x.Action >= uint32(len(scoreboardActionNames)) {
		r.UnknownEnumOption(x.Action, "scoreboard entry action")
		return
	}
	name := scoreboardActionNames[x.Action]
	r.String(&name)
	if name != scoreboardActionNames[x.Action] {
		r.InvalidValue(name, "scoreboard entry action name", "does not match the action")
		return
	}
	r.Varint64(&x.EntryID)
	if x.Action == ScoreboardActionRemove {
		// Removing an entry from every objective at once leaves the name out.
		set := x.ObjectiveName != ""
		r.Bool(&set)
		if set {
			r.String(&x.ObjectiveName)
		}
		return
	}
	r.String(&x.ObjectiveName)
	r.Int32(&x.Score)
	switch x.Action {
	case ScoreboardActionChangePlayer, ScoreboardActionChangeEntity:
		r.Varint64(&x.EntityUniqueID)
	case ScoreboardActionChangeFakePlayer:
		r.String(&x.DisplayName)
	}
}

// ScoreboardIdentityEntry holds an entry to either associate an identity with one of the entries in a
// scoreboard, or to remove associations.
type ScoreboardIdentityEntry struct {
	// EntryID is the unique identifier of the entry that the identity should be associated with, or that
	// associations should be cleared from.
	EntryID int64
	// EntityUniqueID is the unique ID that the entry should be associated with. It is not set if the
	// SetScoreboardIdentity packet is sent to remove associations with identities.
	EntityUniqueID Optional[int64]
}

// Marshal encodes/decodes a ScoreboardIdentityEntry.
func (x *ScoreboardIdentityEntry) Marshal(r IO) {
	r.Varint64(&x.EntryID)
	OptionalFunc(r, &x.EntityUniqueID, r.Varint64)
}
