package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	LessonActionStart = iota
	LessonActionComplete
	LessonActionRestart
)

// LessonProgress is a packet sent by the server to the client to inform the client of updated progress on a lesson.
// This packet only functions on the Minecraft: Education Edition version of the game.
type LessonProgress struct {
	// Identifier is the identifier of the lesson that is being progressed.
	Identifier string
	// Action is the action the client should perform to show progress. This is one of the constants defined above.
	Action uint8
	// Score is the score the client should use when displaying the progress.
	Score int32
}

// ID ...
func (*LessonProgress) ID() uint32 {
	return IDLessonProgress
}

// Marshal ...
func (pk *LessonProgress) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.Action)
	w.Varint32(&pk.Score)
	w.String(&pk.Identifier)
}

// Unmarshal ...
func (pk *LessonProgress) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.Action)
	r.Varint32(&pk.Score)
	r.String(&pk.Identifier)
}
