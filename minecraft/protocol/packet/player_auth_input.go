package packet

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	InputFlagAscend = 1 << iota
	InputFlagDescend
	InputFlagNorthJump
	InputFlagJumpDown
	InputFlagSprintDown
	InputFlagChangeHeight
	InputFlagJumping
	InputFlagAutoJumpingInWater
	InputFlagSneaking
	InputFlagSneakDown
	InputFlagUp
	InputFlagDown
	InputFlagLeft
	InputFlagRight
	InputFlagUpLeft
	InputFlagUpRight
	InputFlagWantUp
	InputFlagWantDown
	InputFlagWantDownSlow
	InputFlagWantUpSlow
	InputFlagSprinting
	InputFlagAscendScaffolding
	InputFlagDescendScaffolding
	InputFlagSneakToggleDown
	InputFlagPersistSneak
)

const (
	InputModeMouse = iota + 1
	InputModeTouch
	InputModeGamePad
	InputModeMotionController
)

const (
	PlayModeNormal = iota
	PlayModeTeaser
	PlayModeScreen
	PlayModeViewer
	PlayModeReality
	PlayModePlacement
	PlayModeLivingRoom
	PlayModeExitLevel
	PlayModeExitLevelLivingRoom
	PlayModeNumModes
)

// PlayerAuthInput is sent by the client to allow for server authoritative movement. It is used to synchronise
// the player input with the position server-side.
// The client sends this packet when the ServerAuthoritativeOverMovement field in the StartGame packet is set
// to true, instead of the MovePlayer packet. The client will send this packet once every tick.
type PlayerAuthInput struct {
	// Pitch and Yaw hold the rotation that the player reports it has.
	Pitch, Yaw float32
	// Position holds the position that the player reports it has.
	Position mgl32.Vec3
	// MoveVector is a Vec2 that specifies the direction in which the player moved, as a combination of X/Z
	// values which are created using the WASD/controller stick state.
	MoveVector mgl32.Vec2
	// HeadYaw is the horizontal rotation of the head that the player reports it has.
	HeadYaw float32
	// InputData is a combination of bit flags that together specify the way the player moved last tick. It
	// is a combination of the flags above.
	InputData uint64
	// InputMode specifies the way that the client inputs data to the screen. It is one of the constants that
	// may be found above.
	InputMode uint32
	// PlayMode specifies the way that the player is playing. The values it holds, which are rather random,
	// may be found above.
	PlayMode uint32
	// GazeDirection is the direction in which the player is gazing, when the PlayMode is PlayModeReality: In
	// other words, when the player is playing in virtual reality.
	GazeDirection mgl32.Vec3
}

// ID ...
func (pk *PlayerAuthInput) ID() uint32 {
	return IDPlayerAuthInput
}

// Marshal ...
func (pk *PlayerAuthInput) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteFloat32(buf, pk.Pitch)
	_ = protocol.WriteFloat32(buf, pk.Yaw)
	_ = protocol.WriteVec3(buf, pk.Position)
	_ = protocol.WriteVec2(buf, pk.MoveVector)
	_ = protocol.WriteFloat32(buf, pk.HeadYaw)
	_ = protocol.WriteVaruint64(buf, pk.InputData)
	_ = protocol.WriteVaruint32(buf, pk.InputMode)
	_ = protocol.WriteVaruint32(buf, pk.PlayMode)
	if pk.PlayMode == PlayModeReality {
		_ = protocol.WriteVec3(buf, pk.GazeDirection)
	}
}

// Unmarshal ...
func (pk *PlayerAuthInput) Unmarshal(buf *bytes.Buffer) error {
	if err := chainErr(
		protocol.Float32(buf, &pk.Yaw),
		protocol.Float32(buf, &pk.Pitch),
		protocol.Vec3(buf, &pk.Position),
		protocol.Vec2(buf, &pk.MoveVector),
		protocol.Float32(buf, &pk.HeadYaw),
		protocol.Varuint64(buf, &pk.InputData),
		protocol.Varuint32(buf, &pk.InputMode),
		protocol.Varuint32(buf, &pk.PlayMode),
	); err != nil {
		return err
	}
	if pk.PlayMode == PlayModeReality {
		return protocol.Vec3(buf, &pk.GazeDirection)
	}
	return nil
}
