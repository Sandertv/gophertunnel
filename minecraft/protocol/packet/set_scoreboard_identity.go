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
	w.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case ScoreboardIdentityActionRegister:
		protocol.Slice(w, &pk.Entries)
	case ScoreboardIdentityActionClear:
		protocol.FuncIOSlice(w, &pk.Entries, protocol.ScoreboardIdentityClearEntry)
	}
}

// Unmarshal ...
func (pk *SetScoreboardIdentity) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.ActionType)
	switch pk.ActionType {
	case ScoreboardIdentityActionRegister:
		protocol.Slice(r, &pk.Entries)
	case ScoreboardIdentityActionClear:
		protocol.FuncIOSlice(r, &pk.Entries, protocol.ScoreboardIdentityClearEntry)
	}
}
