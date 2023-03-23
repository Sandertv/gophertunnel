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

func (pk *GameTestRequest) Marshal(io protocol.IO) {
	io.Varint32(&pk.MaxTestsPerBatch)
	io.Varint32(&pk.Repetitions)
	io.Uint8(&pk.Rotation)
	io.Bool(&pk.StopOnError)
	io.BlockPos(&pk.Position)
	io.Varint32(&pk.TestsPerRow)
	io.String(&pk.Name)
}
