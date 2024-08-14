package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PredictionTypePlayer = iota
	PredictionTypeVehicle
)

// CorrectPlayerMovePrediction is sent by the server if and only if StartGame.ServerAuthoritativeMovementMode
// is set to AuthoritativeMovementModeServerWithRewind. The packet is used to correct movement at a specific
// point in time.
type CorrectPlayerMovePrediction struct {
	// PredictionType is the type of prediction that was corrected. It is one of the constants above.
	PredictionType byte
	// Position is the position that the player is supposed to be at the tick written in the field below.
	// The client will change its current position based on movement after that tick starting from the
	// Position.
	Position mgl32.Vec3
	// Delta is the change in position compared to what the client sent as its position at that specific tick.
	Delta mgl32.Vec3
	// Rotation is the rotation of the player at the tick written in the field below. It is only included if
	// PredictionType is PredictionTypeVehicle.
	Rotation mgl32.Vec2
	// VehicleAngularVelocity is the angular velocity of the vehicle that the rider is riding.
	VehicleAngularVelocity protocol.Optional[float32]
	// OnGround specifies if the player was on the ground at the time of the tick below.
	OnGround bool
	// Tick is the tick of the movement which was corrected by this packet.
	Tick uint64
}

// ID ...
func (*CorrectPlayerMovePrediction) ID() uint32 {
	return IDCorrectPlayerMovePrediction
}

func (pk *CorrectPlayerMovePrediction) Marshal(io protocol.IO) {
	io.Uint8(&pk.PredictionType)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Delta)
	if pk.PredictionType == PredictionTypeVehicle {
		io.Vec2(&pk.Rotation)
		protocol.OptionalFunc(io, &pk.VehicleAngularVelocity, io.Float32)
	}
	io.Bool(&pk.OnGround)
	io.Varuint64(&pk.Tick)
}
