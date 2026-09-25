package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetScore is sent by the server to send the contents of a scoreboard to the player. It may be used to either
// add, remove or edit entries on the scoreboard.
type SetScore struct {
	// Entries is a list of all entries that the client should operate on. Each entry's IdentityType specifies
	// whether it is added, modified or removed.
	Entries []protocol.ScoreboardEntry
}

// ID ...
func (*SetScore) ID() uint32 {
	return IDSetScore
}

func (pk *SetScore) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Entries)
}
