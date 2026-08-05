package packet

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	InputFlagAscend = iota
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
	InputFlagAscendBlock
	InputFlagDescendBlock
	InputFlagSneakToggleDown
	InputFlagPersistSneak
	InputFlagStartSprinting
	InputFlagStopSprinting
	InputFlagStartSneaking
	InputFlagStopSneaking
	InputFlagStartSwimming
	InputFlagStopSwimming
	InputFlagStartJumping
	InputFlagStartGliding
	InputFlagStopGliding
	InputFlagPerformItemInteraction
	InputFlagPerformBlockActions
	InputFlagPerformItemStackRequest
	InputFlagHandledTeleport
	InputFlagEmoting
	InputFlagMissedSwing
	InputFlagStartCrawling
	InputFlagStopCrawling
	InputFlagStartFlying
	InputFlagStopFlying
	InputFlagClientAckServerData
	InputFlagClientPredictedVehicle
	InputFlagPaddlingLeft
	InputFlagPaddlingRight
	InputFlagBlockBreakingDelayEnabled
	InputFlagHorizontalCollision
	InputFlagVerticalCollision
	InputFlagDownLeft
	InputFlagDownRight
	InputFlagStartUsingItem
	InputFlagCameraRelativeMovementEnabled
	InputFlagRotControlledByMoveDirection
	InputFlagStartSpinAttack
	InputFlagStopSpinAttack
	InputFlagIsHotbarTouchOnly
	InputFlagJumpReleasedRaw
	InputFlagJumpPressedRaw
	InputFlagJumpCurrentRaw
	InputFlagSneakReleasedRaw
	InputFlagSneakPressedRaw
	InputFlagSneakCurrentRaw
	InputFlagInternalUpdate

	// InputFlagCount is the number of supported PlayerAuthInput flags.
	InputFlagCount
)

const (
	InputModeUndefined = iota
	InputModeMouse
	InputModeTouch
	InputModeGamePad
	InputModeMotionController
	InputModeCount
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
)

const PlayModeNumModes = PlayModeExitLevelLivingRoom + 1

const (
	InteractionModelTouch = iota
	InteractionModelCrosshair
	InteractionModelClassic
	InteractionModelCount
)

// PlayerAuthInput is sent by the client to allow for server authoritative movement. It is used to synchronise
// the player input with the position server-side.
// The client sends this packet when the ServerAuthoritativeMovementMode field in the StartGame packet is set
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
	// InputData is the set of input flags that together specify the way the player moved last tick. It holds
	// the flags above.
	InputData protocol.InputFlags
	// InputMode specifies the way that the client inputs data to the screen. It is one of the constants that
	// may be found above.
	InputMode uint32
	// PlayMode specifies the way that the player is playing. The values it holds, which are rather random,
	// may be found above.
	PlayMode uint32
	// InteractionModel is a constant representing the interaction model the player is using. It is one of the
	// constants that may be found above.
	InteractionModel int32
	// InteractPitch and interactYaw is the rotation the player is looking that they intend to use for
	// interactions. This is only different to Pitch and Yaw in cases such as VR or when custom cameras
	// being used.
	InteractPitch, InteractYaw float32
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
	// Delta was the delta between the old and the new position. There isn't any practical use for this field
	// as it can be calculated by the server itself.
	Delta mgl32.Vec3
	// ItemInteractionData is the transaction data if the InputData includes an item interaction.
	ItemInteractionData protocol.Optional[protocol.UseItemTransactionData]
	// ItemStackRequest is sent by the client to change an item in their inventory.
	ItemStackRequest protocol.Optional[protocol.ItemStackRequest]
	// BlockActions is a slice of block actions that the client has interacted with.
	BlockActions protocol.Optional[[]protocol.PlayerBlockAction]
	// VehicleRotation is the rotation of the vehicle that the player is in, if any.
	VehicleRotation protocol.Optional[mgl32.Vec2]
	// ClientPredictedVehicle is the unique ID of the vehicle that the client predicts the player to be in.
	ClientPredictedVehicle protocol.Optional[int64]
	// AnalogueMoveVector is a Vec2 that specifies the direction in which the player moved, as a combination
	// of X/Z values which are created using an analogue input.
	AnalogueMoveVector mgl32.Vec2
	// CameraOrientation is the vector that represents the camera's forward direction which can be used to
	// transform movement to be camera relative.
	CameraOrientation mgl32.Vec3
	// RawMoveVector is the value of MoveVector before it is affected by input permissions, sneaking/fly
	// speeds and isn't normalised for analogue inputs.
	RawMoveVector mgl32.Vec2
}

// ID ...
func (pk *PlayerAuthInput) ID() uint32 {
	return IDPlayerAuthInput
}

func (pk *PlayerAuthInput) Marshal(io protocol.IO) {
	io.Float32(&pk.Pitch)
	io.Float32(&pk.Yaw)
	io.Vec3(&pk.Position)
	io.Vec2(&pk.MoveVector)
	io.Float32(&pk.HeadYaw)
	protocol.InputFlagList(io, &pk.InputData, InputFlagCount)
	io.Varuint32(&pk.InputMode)
	io.Varuint32(&pk.PlayMode)
	io.Varint32(&pk.InteractionModel)
	io.Float32(&pk.InteractPitch)
	io.Float32(&pk.InteractYaw)
	io.Varuint64(&pk.Tick)
	io.Vec3(&pk.Delta)
	protocol.DoubleOptionalFunc(io, &pk.ItemInteractionData, io.PlayerInventoryAction)
	protocol.DoubleOptionalFunc(io, &pk.ItemStackRequest, func(x *protocol.ItemStackRequest) {
		x.Marshal(io)
	})
	protocol.DoubleOptionalFunc(io, &pk.BlockActions, func(x *[]protocol.PlayerBlockAction) {
		protocol.Slice(io, x)
	})
	protocol.DoubleOptionalFunc(io, &pk.VehicleRotation, io.Vec2)
	protocol.DoubleOptionalFunc(io, &pk.ClientPredictedVehicle, io.ActorUniqueID)
	io.Vec2(&pk.AnalogueMoveVector)
	io.Vec3(&pk.CameraOrientation)
	io.Vec2(&pk.RawMoveVector)
}
