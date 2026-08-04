package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ScoreboardIdentityActionRegister = iota
	ScoreboardIdentityActionClear
)

// SetScoreboardIdentity is sent by the server to change the identity type of one of the entries on a
// scoreboard. This is used to change, for example, an entry pointing to a player, to a fake player when it
// leaves the server, and to change it back to a real player when it joins again.
// In non-vanilla situations, the packet is quite useless.
type SetScoreboardIdentity struct {
	// ActionType is the type of the action to execute. The action is either ScoreboardIdentityActionRegister
	// to associate an identity with the entry, or ScoreboardIdentityActionClear to remove associations with
	// an entity.
	ActionType byte
	// Entries is a list of all entries in the packet. Each of these entries points to one of the entries on
	// a scoreboard. Depending on ActionType, their identity will either be registered or cleared.
	Entries []protocol.ScoreboardIdentityEntry
}

// ID ...
func (*SetScoreboardIdentity) ID() uint32 {
	return IDSetScoreboardIdentity
}

func (pk *SetScoreboardIdentity) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionType)
	if pk.ActionType != ScoreboardIdentityActionRegister && pk.ActionType != ScoreboardIdentityActionClear {
		io.UnknownEnumOption(pk.ActionType, "scoreboard identity action type")
	}
	protocol.FuncIOSlice(io, &pk.Entries, pk.marshalEntry)
}

func (pk *SetScoreboardIdentity) marshalEntry(io protocol.IO, entry *protocol.ScoreboardIdentityEntry) {
	io.Varint64(&entry.EntryID)
	expectsEntityUniqueID := pk.ActionType == ScoreboardIdentityActionRegister
	hasEntityUniqueID := expectsEntityUniqueID
	io.Bool(&hasEntityUniqueID)
	if hasEntityUniqueID != expectsEntityUniqueID {
		io.InvalidValue(hasEntityUniqueID, "scoreboard identity entity unique ID presence", "does not match action type")
	}
	if hasEntityUniqueID {
		io.Varint64(&entry.EntityUniqueID)
	} else {
		entry.EntityUniqueID = 0
	}
}
