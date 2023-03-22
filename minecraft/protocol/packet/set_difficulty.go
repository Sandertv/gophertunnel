package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SetDifficulty is sent by the server to update the client-side difficulty of the client. The actual effect
// of this packet on the client isn't very significant, as the difficulty is handled server-side.
type SetDifficulty struct {
	// Difficulty is the new difficulty that the world has.
	Difficulty uint32
}

// ID ...
func (*SetDifficulty) ID() uint32 {
	return IDSetDifficulty
}

// Marshal ...
func (pk *SetDifficulty) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SetDifficulty) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SetDifficulty) marshal(r protocol.IO) {
	r.Varuint32(&pk.Difficulty)
}
