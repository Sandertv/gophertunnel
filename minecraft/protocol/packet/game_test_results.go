package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// GameTestResults is a packet sent in response to the GameTestRequest packet, with a boolean indicating whether the
// test was successful or not, and an error string if the test failed.
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

func (pk *GameTestResults) Marshal(io protocol.IO) {
	io.Bool(&pk.Succeeded)
	io.String(&pk.Error)
	io.String(&pk.Name)
}
