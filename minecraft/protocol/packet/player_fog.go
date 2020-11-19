package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type PlayerFog struct {
	Stack []string
}

// ID ...
func (PlayerFog) ID() uint32 {
	return IDPlayerFog
}

// Marshal ...
func (pk PlayerFog) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Stack))
	w.Varuint32(&l)
	for i := range pk.Stack {
		w.String(&pk.Stack[i])
	}
}

// Unmarshal ...
func (pk PlayerFog) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)
	pk.Stack = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Stack[i])
	}
}
