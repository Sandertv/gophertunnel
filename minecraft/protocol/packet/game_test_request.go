package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	GameTestRequestRotation0 = iota
	GameTestRequestRotation90
	GameTestRequestRotation180
	GameTestRequestRotation270
	GameTestRequestRotation360
)

// GameTestRequest ...
// TODO: Document this.
type GameTestRequest struct {
	// Name represents the name of the test.
	Name string
	// Rotation represents the rotation of the test. It is one of the constants above.
	Rotation uint8
	// Repetitions represents the amount of times the test will be run.
	Repetitions int32
	// Position is the position at which the test will be performed.
	Position protocol.BlockPos
	// StopOnError indicates whether the test should immediately stop when an error is encountered.
	StopOnError bool
	// TestsPerRow ...
	TestsPerRow int32
	// MaxTestsPerBatch ...
	MaxTestsPerBatch int32
}

// ID ...
func (pk *GameTestRequest) ID() uint32 {
	return IDGameTestRequest
}

// Marshal ...
func (pk *GameTestRequest) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.MaxTestsPerBatch)
	w.Varint32(&pk.Repetitions)
	w.Uint8(&pk.Rotation)
	w.Bool(&pk.StopOnError)
	w.BlockPos(&pk.Position)
	w.Varint32(&pk.TestsPerRow)
	w.String(&pk.Name)
}

// Unmarshal ...
func (pk *GameTestRequest) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.MaxTestsPerBatch)
	r.Varint32(&pk.Repetitions)
	r.Uint8(&pk.Rotation)
	r.Bool(&pk.StopOnError)
	r.BlockPos(&pk.Position)
	r.Varint32(&pk.TestsPerRow)
	r.String(&pk.Name)
}
