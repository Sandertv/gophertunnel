package protocol

const (
	ScoreboardIdentityRemove = iota
	ScoreboardIdentityPlayer
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

// Marshal encodes/decodes a ScoreboardEntry x as an entry for modification.
func (x *ScoreboardEntry) Marshal(r IO) {
	variant := uint32(x.IdentityType)
	r.Varuint32(&variant)
	x.IdentityType = byte(variant)

	typeNames := [...]string{"remove", "changeplayer", "changeentity", "changefakeplayer"}
	if variant >= uint32(len(typeNames)) {
		r.UnknownEnumOption(variant, "scoreboard entry variant")
		return
	}
	typeName := typeNames[variant]
	r.String(&typeName)
	r.Varint64(&x.EntryID)
	switch x.IdentityType {
	case ScoreboardIdentityRemove:
		objective := Optional[string]{}
		if x.ObjectiveName != "" {
			objective = Option(x.ObjectiveName)
		}
		OptionalFunc(r, &objective, r.String)
		x.ObjectiveName, _ = objective.Value()
	case ScoreboardIdentityEntity, ScoreboardIdentityPlayer:
		r.String(&x.ObjectiveName)
		r.Int32(&x.Score)
		r.ActorUniqueID(&x.EntityUniqueID)
	case ScoreboardIdentityFakePlayer:
		r.String(&x.ObjectiveName)
		r.Int32(&x.Score)
		r.String(&x.DisplayName)
	}
}

// ScoreboardIdentityEntry holds an entry to either associate an identity with one of the entries in a
// scoreboard, or to remove associations.
type ScoreboardIdentityEntry struct {
	// EntryID is the unique identifier of the entry that the identity should be associated with, or that
	// associations should be cleared from.
	EntryID int64
	// EntityUniqueID is the optional unique ID that the entry should be associated with.
	EntityUniqueID Optional[int64]
}

// Marshal encodes/decodes a ScoreboardIdentityEntry.
func (x *ScoreboardIdentityEntry) Marshal(r IO) {
	r.Varint64(&x.EntryID)
	OptionalFunc(r, &x.EntityUniqueID, r.ActorUniqueID)
}
