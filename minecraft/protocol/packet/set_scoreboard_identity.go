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

// Marshal ...
func (pk *SetScoreboardIdentity) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Entries))
	w.Uint8(&pk.ActionType)
	w.Varuint32(&l)
	for _, entry := range pk.Entries {
		w.Varint64(&entry.EntryID)
		if pk.ActionType == ScoreboardIdentityActionRegister {
			w.Varint64(&entry.EntityUniqueID)
		}
	}
}

// Unmarshal ...
func (pk *SetScoreboardIdentity) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Uint8(&pk.ActionType)
	r.Varuint32(&count)

	pk.Entries = make([]protocol.ScoreboardIdentityEntry, count)
	for i := uint32(0); i < count; i++ {
		r.Varint64(&pk.Entries[i].EntryID)
		if pk.ActionType == ScoreboardIdentityActionRegister {
			r.Varint64(&pk.Entries[i].EntityUniqueID)
		}
	}
}
