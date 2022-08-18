package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GameTestResults ...
// TODO: Document this.
type GameTestResults struct {
	// Name represents the name of the test.
	Name string
	// Succeeded indicates whether the test succeeded or not.
	Succeeded bool
	// Error is the error that occurred. If Succeeded is true, this field is empty.
	Error string
}

// ID ...
func (pk *GameTestResults) ID() uint32 {
	return IDGameTestResults
}

// Marshal ...
func (pk *GameTestResults) Marshal(w *protocol.Writer) {
	w.Bool(&pk.Succeeded)
	w.String(&pk.Error)
	w.String(&pk.Name)
}

// Unmarshal ...
func (pk *GameTestResults) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.Succeeded)
	r.String(&pk.Error)
	r.String(&pk.Name)
}
