package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AnimateEntity is sent by the server to animate an entity client-side. It may be used to play a single
// animation, or to activate a controller which can start a sequence of animations based on different
// conditions specified in an animation controller.
// Much of the documentation of this packet can be found at
// https://learn.microsoft.com/en-us/minecraft/creator/reference/content/animationsreference
type AnimateEntity struct {
	// Animation is the name of a single animation to start playing.
	Animation string
	// NextState is the first state to start with. These states are declared in animation controllers (which,
	// in themselves, are animations too). These states in turn may have animations and transitions to move to
	// a next state.
	NextState string
	// StopCondition is a MoLang expression that specifies when the animation should be stopped.
	StopCondition string
	// StopConditionVersion is the MoLang stop condition version.
	StopConditionVersion int32
	// Controller is the animation controller that is used to manage animations. These controllers decide when
	// to play which animation.
	Controller string
	// BlendOutTime does not currently seem to be used.
	BlendOutTime float32
	// EntityRuntimeIDs is list of runtime IDs of entities that the animation should be applied to.
	EntityRuntimeIDs []uint64
}

// ID ...
func (*AnimateEntity) ID() uint32 {
	return IDAnimateEntity
}

func (pk *AnimateEntity) Marshal(io protocol.IO) {
	io.String(&pk.Animation)
	io.String(&pk.NextState)
	io.String(&pk.StopCondition)
	io.Int32(&pk.StopConditionVersion)
	io.String(&pk.Controller)
	io.Float32(&pk.BlendOutTime)
	protocol.FuncSlice(io, &pk.EntityRuntimeIDs, io.Varuint64)
}
