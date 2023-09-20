package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
)

const (
	AudioListenerCamera = iota
	AudioListenerPlayer
)

const (
	EasingTypeLinear = iota
	EasingTypeSpring
	EasingTypeInQuad
	EasingTypeOutQuad
	EasingTypeInOutQuad
	EasingTypeInCubic
	EasingTypeOutCubic
	EasingTypeInOutCubic
	EasingTypeInQuart
	EasingTypeOutQuart
	EasingTypeInOutQuart
	EasingTypeInQuint
	EasingTypeOutQuint
	EasingTypeInOutQuint
	EasingTypeInSine
	EasingTypeOutSine
	EasingTypeInOutSine
	EasingTypeInExpo
	EasingTypeOutExpo
	EasingTypeInOutExpo
	EasingTypeInCirc
	EasingTypeOutCirc
	EasingTypeInOutCirc
	EasingTypeInBounce
	EasingTypeOutBounce
	EasingTypeInOutBounce
	EasingTypeInBack
	EasingTypeOutBack
	EasingTypeInOutBack
	EasingTypeInElastic
	EasingTypeOutElastic
	EasingTypeInOutElastic
)

// CameraEase represents an easing function that can be used by a CameraInstructionSet.
type CameraEase struct {
	// Type is the type of easing function used. This is one of the constants above.
	Type uint8
	// Duration is the time in seconds that the easing function should take.
	Duration float32
}

// Marshal encodes/decodes a CameraEase.
func (x *CameraEase) Marshal(r IO) {
	r.Uint8(&x.Type)
	r.Float32(&x.Duration)
}

// CameraInstruction represents an instruction that can be one of three different operations.
type CameraInstruction struct {
	// Set is a camera instruction that sets the camera to a specified preset.
	Set Optional[CameraInstructionSet]
	// Clear can be set to true to clear all the current camera instructions.
	Clear Optional[bool]
	// Fade is a camera instruction that fades the screen to a specified colour.
	Fade Optional[CameraInstructionFade]
}

// Marshal encodes/decodes a CameraInstruction.
func (x *CameraInstruction) Marshal(r IO) {
	OptionalMarshaler(r, &x.Set)
	OptionalFunc(r, &x.Clear, r.Bool)
	OptionalMarshaler(r, &x.Fade)
}

// CameraInstructionSet represents a camera instruction that sets the camera to a specified preset and can be extended
// with easing functions and translations to the camera's position and rotation.
type CameraInstructionSet struct {
	// Preset is the index of the preset in the CameraPresets packet sent to the player.
	Preset uint32
	// Ease represents the easing function that is used by the instruction.
	Ease Optional[CameraEase]
	// Position represents the position of the camera.
	Position Optional[mgl32.Vec3]
	// Rotation represents the rotation of the camera.
	Rotation Optional[mgl32.Vec2]
	// Facing is a vector that the camera will always face towards during the duration of the instruction.
	Facing Optional[mgl32.Vec3]
	// Default determines whether the camera is a default camera or not.
	Default Optional[bool]
}

// Marshal encodes/decodes a CameraInstructionSet.
func (x *CameraInstructionSet) Marshal(r IO) {
	r.Uint32(&x.Preset)
	OptionalMarshaler(r, &x.Ease)
	OptionalFunc(r, &x.Position, r.Vec3)
	OptionalFunc(r, &x.Rotation, r.Vec2)
	OptionalFunc(r, &x.Facing, r.Vec3)
	OptionalFunc(r, &x.Default, r.Bool)
}

// CameraInstructionFade represents a camera instruction that fades the screen to a specified colour.
type CameraInstructionFade struct {
	// FadeInDuration is the time in seconds for the screen to fully fade in.
	FadeInDuration float32
	// WaitDuration is time in seconds to wait before fading out.
	WaitDuration float32
	// FadeOutDuration is the time in seconds for the screen to fully fade out.
	FadeOutDuration float32
	// Colour is the colour of the screen to fade to. This only uses the red, green and blue components.
	Colour color.RGBA
}

// Marshal encodes/decodes a CameraInstructionFade.
func (x *CameraInstructionFade) Marshal(r IO) {
	r.Float32(&x.FadeInDuration)
	r.Float32(&x.WaitDuration)
	r.Float32(&x.FadeOutDuration)
	r.RGB(&x.Colour)
}

// CameraPreset represents a basic preset that can be extended upon by more complex instructions.
type CameraPreset struct {
	// Name is the name of the preset. Each preset must have their own unique name.
	Name string
	// Parent is the name of the preset that this preset extends upon. This can be left empty.
	Parent string
	// PosX is the default X position of the camera.
	PosX Optional[float32]
	// PosY is the default Y position of the camera.
	PosY Optional[float32]
	// PosZ is the default Z position of the camera.
	PosZ Optional[float32]
	// RotX is the default pitch of the camera.
	RotX Optional[float32]
	// RotY is the default yaw of the camera.
	RotY Optional[float32]
	// AudioListener defines where the audio should be played from when using this preset. This is one of the constants
	// above.
	AudioListener Optional[byte]
	// PlayerEffects is currently unknown.
	PlayerEffects Optional[bool]
}

// Marshal encodes/decodes a CameraPreset.
func (x *CameraPreset) Marshal(r IO) {
	r.String(&x.Name)
	r.String(&x.Parent)
	OptionalFunc(r, &x.PosX, r.Float32)
	OptionalFunc(r, &x.PosY, r.Float32)
	OptionalFunc(r, &x.PosZ, r.Float32)
	OptionalFunc(r, &x.RotX, r.Float32)
	OptionalFunc(r, &x.RotY, r.Float32)
	OptionalFunc(r, &x.AudioListener, r.Uint8)
	OptionalFunc(r, &x.PlayerEffects, r.Bool)
}
