package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"image/color"
)

const (
	AimAssistTargetModeAngle = iota
	AimAssistTargetModeDistance
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
	// ViewOffset is an offset based on a pivot point to the player, causing the camera to be shifted in a
	// certain direction.
	ViewOffset Optional[mgl32.Vec2]
	// EntityOffset is an offset from the entity that the camera should be rendered at.
	EntityOffset Optional[mgl32.Vec3]
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
	OptionalFunc(r, &x.ViewOffset, r.Vec2)
	OptionalFunc(r, &x.EntityOffset, r.Vec3)
	OptionalFunc(r, &x.Default, r.Bool)
}

// CameraFadeTimeData represents the time data for a CameraInstructionFade.
type CameraFadeTimeData struct {
	// FadeInDuration is the time in seconds for the screen to fully fade in.
	FadeInDuration float32
	// WaitDuration is time in seconds to wait before fading out.
	WaitDuration float32
	// FadeOutDuration is the time in seconds for the screen to fully fade out.
	FadeOutDuration float32
}

// Marshal encodes/decodes a CameraFadeTimeData.
func (x *CameraFadeTimeData) Marshal(r IO) {
	r.Float32(&x.FadeInDuration)
	r.Float32(&x.WaitDuration)
	r.Float32(&x.FadeOutDuration)
}

// CameraInstructionFade represents a camera instruction that fades the screen to a specified colour.
type CameraInstructionFade struct {
	// TimeData is the time data for the fade, which includes the fade in duration, wait duration and fade out
	// duration.
	TimeData Optional[CameraFadeTimeData]
	// Colour is the colour of the screen to fade to. This only uses the red, green and blue components.
	Colour Optional[color.RGBA]
}

// Marshal encodes/decodes a CameraInstructionFade.
func (x *CameraInstructionFade) Marshal(r IO) {
	OptionalMarshaler(r, &x.TimeData)
	OptionalFunc(r, &x.Colour, r.RGB)
}

// CameraInstructionTarget represents a camera instruction that targets a specific entity.
type CameraInstructionTarget struct {
	// CenterOffset is the offset from the center of the entity that the camera should target.
	CenterOffset Optional[mgl32.Vec3]
	// EntityUniqueID is the unique ID of the entity that the camera should target.
	EntityUniqueID int64
}

// Marshal encodes/decodes a CameraInstructionTarget.
func (x *CameraInstructionTarget) Marshal(r IO) {
	OptionalFunc(r, &x.CenterOffset, r.Vec3)
	r.Int64(&x.EntityUniqueID)
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
	// RotationSpeed is the speed at which the camera should rotate.
	RotationSpeed Optional[float32]
	// SnapToTarget determines whether the camera should snap to the target entity or not.
	SnapToTarget Optional[bool]
	// HorizontalRotationLimit is the horizontal rotation limit of the camera.
	HorizontalRotationLimit Optional[mgl32.Vec2]
	// VerticalRotationLimit is the vertical rotation limit of the camera.
	VerticalRotationLimit Optional[mgl32.Vec2]
	// ContinueTargeting determines whether the camera should continue targeting the entity or not.
	ContinueTargeting Optional[bool]
	// BlockListeningRadius ...
	BlockListeningRadius Optional[float32]
	// ViewOffset is only used in a follow_orbit camera and controls an offset based on a pivot point to the
	// player, causing it to be shifted in a certain direction.
	ViewOffset Optional[mgl32.Vec2]
	// EntityOffset controls the offset from the entity that the camera should be rendered at.
	EntityOffset Optional[mgl32.Vec3]
	// Radius is only used in a follow_orbit camera and controls how far away from the player the camera should
	// be rendered.
	Radius Optional[float32]
	// AudioListener defines where the audio should be played from when using this preset. This is one of the
	// constants above.
	AudioListener Optional[byte]
	// PlayerEffects is currently unknown.
	PlayerEffects Optional[bool]
	// AlignTargetAndCameraForward determines whether the camera should align the target and the camera forward
	// or not.
	AlignTargetAndCameraForward Optional[bool]
	// AimAssist ...
	AimAssist Optional[CameraPresetAimAssist]
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
	OptionalFunc(r, &x.RotationSpeed, r.Float32)
	OptionalFunc(r, &x.SnapToTarget, r.Bool)
	OptionalFunc(r, &x.HorizontalRotationLimit, r.Vec2)
	OptionalFunc(r, &x.VerticalRotationLimit, r.Vec2)
	OptionalFunc(r, &x.ContinueTargeting, r.Bool)
	OptionalFunc(r, &x.ViewOffset, r.Vec2)
	OptionalFunc(r, &x.EntityOffset, r.Vec3)
	OptionalFunc(r, &x.Radius, r.Float32)
	OptionalFunc(r, &x.AudioListener, r.Uint8)
	OptionalFunc(r, &x.PlayerEffects, r.Bool)
	OptionalFunc(r, &x.AlignTargetAndCameraForward, r.Bool)
}

type CameraPresetAimAssist struct {
	PresetID   Optional[string]
	TargetMode Optional[int32]
	Angle      Optional[mgl32.Vec2]
	Distance   Optional[float32]
}

// Marshal encodes/decodes a CameraPresetAimAssist.
func (x *CameraPresetAimAssist) Marshal(r IO) {
	OptionalFunc(r, &x.PresetID, r.String)
	OptionalFunc(r, &x.TargetMode, r.Int32)
	OptionalFunc(r, &x.Angle, r.Vec2)
	OptionalFunc(r, &x.Distance, r.Float32)
}

type CameraAimAssistCategories struct {
	Identifier string
	Categories []CameraAimAssistCategory
}

// Marshal encodes/decodes a CameraAimAssistCategories.
func (x *CameraAimAssistCategories) Marshal(r IO) {
	r.String(&x.Identifier)
	Slice(r, &x.Categories)
}

type CameraAimAssistCategory struct {
	Name       string
	Priorities CameraAimAssistPriorities
}

// Marshal encodes/decodes a CameraAimAssistCategory.
func (x *CameraAimAssistCategory) Marshal(r IO) {
	r.String(&x.Name)
	Single(r, &x.Priorities)
}

type CameraAimAssistPriorities struct {
	Entities      []CameraAimAssistPriority
	Blocks        []CameraAimAssistPriority
	EntityDefault Optional[int32]
	BlockDefault  Optional[int32]
}

// Marshal encodes/decodes a CameraAimAssistPriorities.
func (x *CameraAimAssistPriorities) Marshal(r IO) {
	Slice(r, &x.Entities)
	Slice(r, &x.Blocks)
	OptionalFunc(r, &x.EntityDefault, r.Int32)
	OptionalFunc(r, &x.BlockDefault, r.Int32)
}

type CameraAimAssistPriority struct {
	ID       string
	Priority int32
}

// Marshal encodes/decodes a CameraAimAssistPriority.
func (x *CameraAimAssistPriority) Marshal(r IO) {
	r.String(&x.ID)
	r.Int32(&x.Priority)
}

type CameraAimAssistPreset struct {
	Identifier          string
	Categories          string
	BlockExclusions     []string
	LiquidTargets       []string
	ItemSettings        []CameraAimAssistItemSettings
	DefaultItemSettings Optional[string]
	HandSettings        Optional[string]
}

// Marshal encodes/decodes a CameraAimAssistPreset.
func (x *CameraAimAssistPreset) Marshal(r IO) {
	r.String(&x.Identifier)
	r.String(&x.Categories)
	FuncSlice(r, &x.BlockExclusions, r.String)
	FuncSlice(r, &x.LiquidTargets, r.String)
	Slice(r, &x.ItemSettings)
	OptionalFunc(r, &x.DefaultItemSettings, r.String)
	OptionalFunc(r, &x.HandSettings, r.String)
}

type CameraAimAssistItemSettings struct {
	ItemID   string
	Category string
}

// Marshal encodes/decodes a CameraAimAssistItemSettings.
func (x *CameraAimAssistItemSettings) Marshal(r IO) {
	r.String(&x.ItemID)
	r.String(&x.Category)
}
