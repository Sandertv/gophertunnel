package packet

import (
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
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
	// Delta was the delta between the old and the new position. There isn't any practical use for this field
	// as it can be calculated by the server itself.
	Delta mgl32.Vec3
	// ItemInteractionData ...
	ItemInteractionData protocol.UseItemTransactionData
	// ItemStackRequest ...
	ItemStackRequest protocol.ItemStackRequest
	// BlockActions ...
	BlockActions []protocol.PlayerBlockAction
}

// ID ...
func (pk *PlayerAuthInput) ID() uint32 {
	return IDPlayerAuthInput
}

// Marshal ...
func (pk *PlayerAuthInput) Marshal(w *protocol.Writer) {
	w.Float32(&pk.Pitch)
	w.Float32(&pk.Yaw)
	w.Vec3(&pk.Position)
	w.Vec2(&pk.MoveVector)
	w.Float32(&pk.HeadYaw)
	w.Varuint64(&pk.InputData)
	w.Varuint32(&pk.InputMode)
	w.Varuint32(&pk.PlayMode)
	if pk.PlayMode == PlayModeReality {
		w.Vec3(&pk.GazeDirection)
	}
	w.Varuint64(&pk.Tick)
	w.Vec3(&pk.Delta)

	if pk.InputData&InputFlagPerformItemInteraction != 0 {
		w.Varint32(&pk.ItemInteractionData.LegacyRequestID)
		if pk.ItemInteractionData.LegacyRequestID != 0 {
			l := uint32(len(pk.ItemInteractionData.LegacySetItemSlots))
			w.Varuint32(&l)
			for _, slot := range pk.ItemInteractionData.LegacySetItemSlots {
				protocol.SetItemSlot(w, &slot)
			}
		}
		w.Bool(&pk.ItemInteractionData.HasNetworkIDs)
		l := uint32(len(pk.ItemInteractionData.Actions))
		w.Varuint32(&l)
		for _, a := range pk.ItemInteractionData.Actions {
			protocol.InvAction(w, &a, pk.ItemInteractionData.HasNetworkIDs)
		}
		w.Varuint32(&pk.ItemInteractionData.ActionType)
		w.BlockPos(&pk.ItemInteractionData.BlockPosition)
		w.Varint32(&pk.ItemInteractionData.BlockFace)
		w.Varint32(&pk.ItemInteractionData.HotBarSlot)
		w.Item(&pk.ItemInteractionData.HeldItem)
		w.Vec3(&pk.ItemInteractionData.Position)
		w.Vec3(&pk.ItemInteractionData.ClickedPosition)
		w.Varuint32(&pk.ItemInteractionData.BlockRuntimeID)
	}

	if pk.InputData&InputFlagPerformItemStackRequest != 0 {
		protocol.WriteStackRequest(w, &pk.ItemStackRequest)
	}

	if pk.InputData&InputFlagPerformBlockActions != 0 {
		l := int32(len(pk.BlockActions))
		w.Varint32(&l)
		for _, action := range pk.BlockActions {
			protocol.BlockAction(w, &action)
		}
	}
}

// Unmarshal ...
func (pk *PlayerAuthInput) Unmarshal(r *protocol.Reader) {
	r.Float32(&pk.Pitch)
	r.Float32(&pk.Yaw)
	r.Vec3(&pk.Position)
	r.Vec2(&pk.MoveVector)
	r.Float32(&pk.HeadYaw)
	r.Varuint64(&pk.InputData)
	r.Varuint32(&pk.InputMode)
	r.Varuint32(&pk.PlayMode)
	if pk.PlayMode == PlayModeReality {
		r.Vec3(&pk.GazeDirection)
	}
	r.Varuint64(&pk.Tick)
	r.Vec3(&pk.Delta)

	if pk.InputData&InputFlagPerformItemInteraction != 0 {
		r.Varint32(&pk.ItemInteractionData.LegacyRequestID)
		if pk.ItemInteractionData.LegacyRequestID != 0 {
			l := uint32(len(pk.ItemInteractionData.LegacySetItemSlots))
			r.Varuint32(&l)
			for _, slot := range pk.ItemInteractionData.LegacySetItemSlots {
				protocol.SetItemSlot(r, &slot)
			}
		}
		r.Bool(&pk.ItemInteractionData.HasNetworkIDs)
		var l uint32
		r.Varuint32(&l)
		pk.ItemInteractionData.Actions = make([]protocol.InventoryAction, l)
		for i := uint32(0); i < l; i++ {
			protocol.InvAction(r, &pk.ItemInteractionData.Actions[i], pk.ItemInteractionData.HasNetworkIDs)
		}
		r.Varuint32(&pk.ItemInteractionData.ActionType)
		r.BlockPos(&pk.ItemInteractionData.BlockPosition)
		r.Varint32(&pk.ItemInteractionData.BlockFace)
		r.Varint32(&pk.ItemInteractionData.HotBarSlot)
		r.Item(&pk.ItemInteractionData.HeldItem)
		r.Vec3(&pk.ItemInteractionData.Position)
		r.Vec3(&pk.ItemInteractionData.ClickedPosition)
		r.Varuint32(&pk.ItemInteractionData.BlockRuntimeID)
	}

	if pk.InputData&InputFlagPerformItemStackRequest != 0 {
		protocol.StackRequest(r, &pk.ItemStackRequest)
	}

	if pk.InputData&InputFlagPerformBlockActions != 0 {
		var l int32
		r.Varint32(&l)
		pk.BlockActions = make([]protocol.PlayerBlockAction, l)
		for i := int32(0); i < l; i++ {
			protocol.BlockAction(r, &pk.BlockActions[i])
		}
	}
}
