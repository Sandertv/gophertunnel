package protocol

const (
	AttributeLayerPayloadTypeUpdateLayers = iota
	AttributeLayerPayloadTypeUpdateSettings
	AttributeLayerPayloadTypeUpdateEnvironment
	AttributeLayerPayloadTypeRemoveEnvironment
)

const (
	AttributeDataTypeBool = iota
	AttributeDataTypeFloat
	AttributeDataTypeColour
)

const (
	AttributeBoolOperationOverride = iota
	AttributeBoolOperationAlphaBlend
	AttributeBoolOperationAnd
	AttributeBoolOperationNand
	AttributeBoolOperationOr
	AttributeBoolOperationNor
	AttributeBoolOperationXor
	AttributeBoolOperationXnor
)

const (
	AttributeFloatOperationOverride = iota
	AttributeFloatOperationAlphaBlend
	AttributeFloatOperationAdd
	AttributeFloatOperationSubtract
	AttributeFloatOperationMultiply
	AttributeFloatOperationMinimum
	AttributeFloatOperationMaximum
)

const (
	AttributeColourOperationOverride = iota
	AttributeColourOperationAlphaBlend
	AttributeColourOperationAdd
	AttributeColourOperationSubtract
	AttributeColourOperationMultiply
)

// AttributeData represents a polymorphic attribute value.
type AttributeData struct {
	// Type is the attribute data type. It is one of the AttributeDataType constants.
	Type uint32
	// BoolValue is the boolean value if Type is AttributeDataTypeBool.
	BoolValue bool
	// BoolOperation is the optional operation for boolean attributes.
	BoolOperation Optional[int32]
	// FloatValue is the float value if Type is AttributeDataTypeFloat.
	FloatValue float32
	// FloatOperation is the optional operation for float attributes.
	FloatOperation Optional[int32]
	// FloatConstraintMin is the optional minimum constraint for float attributes.
	FloatConstraintMin Optional[float32]
	// FloatConstraintMax is the optional maximum constraint for float attributes.
	FloatConstraintMax Optional[float32]
	// ColourValue is the colour value if Type is AttributeDataTypeColour.
	ColourValue int32
	// ColourOperation is the optional operation for colour attributes.
	ColourOperation Optional[int32]
}

// Marshal encodes/decodes an AttributeData.
func (x *AttributeData) Marshal(r IO) {
	r.Varuint32(&x.Type)
	switch x.Type {
	case AttributeDataTypeBool:
		r.Bool(&x.BoolValue)
		OptionalFunc(r, &x.BoolOperation, r.Int32)
	case AttributeDataTypeFloat:
		r.Float32(&x.FloatValue)
		OptionalFunc(r, &x.FloatOperation, r.Int32)
		OptionalFunc(r, &x.FloatConstraintMin, r.Float32)
		OptionalFunc(r, &x.FloatConstraintMax, r.Float32)
	case AttributeDataTypeColour:
		r.Int32(&x.ColourValue)
		OptionalFunc(r, &x.ColourOperation, r.Int32)
	default:
		r.UnknownEnumOption(x.Type, "attribute data type")
	}
}

const (
	NoiseAlignmentTypeMinLocalTransitionEnd = iota
)

// NoiseAlignment represents the way the noise of an environment attribute transition is aligned.
type NoiseAlignment struct {
	// Type is the type of the alignment. It is one of the NoiseAlignmentType constants above.
	Type byte
	// Value is the value that the noise is aligned against, the meaning of which depends on Type.
	Value uint32
}

// Marshal encodes/decodes a NoiseAlignment.
func (x *NoiseAlignment) Marshal(r IO) {
	r.Uint8(&x.Type)
	r.Varuint32(&x.Value)
}

const (
	EnvironmentAttributePayloadTypeConstant = iota
	EnvironmentAttributePayloadTypeTransition
	EnvironmentAttributePayloadTypeNoiseTransition
)

// AttributeTransitionSettings holds the settings of a transition between two environment attribute values.
type AttributeTransitionSettings struct {
	// TotalTransitionTicks is the total number of ticks for the transition.
	TotalTransitionTicks uint32
	// CurrentTransitionTicks is the number of ticks elapsed in the current transition.
	CurrentTransitionTicks uint32
	// EaseType is the easing function used for the transition. It is one of the EasingType constants.
	EaseType int32
	// ClockName is the name of the world clock that drives the transition.
	ClockName string
}

// Marshal encodes/decodes an AttributeTransitionSettings.
func (x *AttributeTransitionSettings) Marshal(r IO) {
	r.Varuint32(&x.TotalTransitionTicks)
	r.Varuint32(&x.CurrentTransitionTicks)
	r.Varint32(&x.EaseType)
	r.String(&x.ClockName)
}

// AttributeNoiseTransitionSettings holds the settings of a noise based transition between two environment
// attribute values.
type AttributeNoiseTransitionSettings struct {
	// TotalTransitionTicks is the total number of ticks for the transition.
	TotalTransitionTicks uint32
	// CurrentTransitionTicks is the number of ticks elapsed in the current transition.
	CurrentTransitionTicks uint32
	// EaseType is the easing function used for the transition. It is one of the EasingType constants.
	EaseType int32
	// ClockName is the name of the world clock that drives the transition.
	ClockName string
	// LocalTransitionTicks is the number of ticks elapsed in the local transition.
	LocalTransitionTicks uint32
	// NoiseName is the name of the noise used by the transition.
	NoiseName string
	// NoiseAlignment is the alignment of the noise used by the transition.
	NoiseAlignment NoiseAlignment
}

// Marshal encodes/decodes an AttributeNoiseTransitionSettings.
func (x *AttributeNoiseTransitionSettings) Marshal(r IO) {
	r.Varuint32(&x.TotalTransitionTicks)
	r.Varuint32(&x.CurrentTransitionTicks)
	r.Varint32(&x.EaseType)
	r.String(&x.ClockName)
	r.Varuint32(&x.LocalTransitionTicks)
	r.String(&x.NoiseName)
	Single(r, &x.NoiseAlignment)
}

// EnvironmentAttributeData represents an environment attribute that either holds a constant value or
// transitions between two values.
type EnvironmentAttributeData struct {
	// AttributeName is the name of the attribute.
	AttributeName string
	// PayloadType is the type of the payload of the attribute. It is one of the EnvironmentAttributePayloadType
	// constants above.
	PayloadType uint32
	// Attribute is the constant attribute value. It is used if PayloadType is
	// EnvironmentAttributePayloadTypeConstant.
	Attribute AttributeData
	// FromAttribute is the starting attribute of the transition. It is used if PayloadType is
	// EnvironmentAttributePayloadTypeTransition or EnvironmentAttributePayloadTypeNoiseTransition.
	FromAttribute AttributeData
	// ToAttribute is the target attribute of the transition. It is used if PayloadType is
	// EnvironmentAttributePayloadTypeTransition or EnvironmentAttributePayloadTypeNoiseTransition.
	ToAttribute AttributeData
	// TransitionSettings holds the settings of the transition. It is used if PayloadType is
	// EnvironmentAttributePayloadTypeTransition.
	TransitionSettings AttributeTransitionSettings
	// NoiseTransitionSettings holds the settings of the noise transition. It is used if PayloadType is
	// EnvironmentAttributePayloadTypeNoiseTransition.
	NoiseTransitionSettings AttributeNoiseTransitionSettings
}

// Marshal encodes/decodes an EnvironmentAttributeData.
func (x *EnvironmentAttributeData) Marshal(r IO) {
	r.String(&x.AttributeName)
	r.Varuint32(&x.PayloadType)
	switch x.PayloadType {
	case EnvironmentAttributePayloadTypeConstant:
		Single(r, &x.Attribute)
	case EnvironmentAttributePayloadTypeTransition:
		Single(r, &x.FromAttribute)
		Single(r, &x.ToAttribute)
		Single(r, &x.TransitionSettings)
	case EnvironmentAttributePayloadTypeNoiseTransition:
		Single(r, &x.FromAttribute)
		Single(r, &x.ToAttribute)
		Single(r, &x.NoiseTransitionSettings)
	default:
		r.UnknownEnumOption(x.PayloadType, "environment attribute payload type")
	}
}

// AttributeLayerSettings represents settings for an attribute layer.
type AttributeLayerSettings struct {
	// Priority is the priority of the layer.
	Priority int32
	// FloatWeight is the weight of the layer.
	FloatWeight float32
	// Enabled indicates if the layer is enabled.
	Enabled bool
	// TransitionsPaused indicates if transitions are paused for this layer.
	TransitionsPaused bool
}

// Marshal encodes/decodes an AttributeLayerSettings.
func (x *AttributeLayerSettings) Marshal(r IO) {
	r.Int32(&x.Priority)
	r.Float32(&x.FloatWeight)
	r.Bool(&x.Enabled)
	r.Bool(&x.TransitionsPaused)
}

// AttributeLayerData represents a complete attribute layer.
type AttributeLayerData struct {
	// Name is the name of the attribute layer.
	Name string
	// DimensionID is the dimension the layer applies to.
	DimensionID int32
	// Settings is the layer's settings.
	Settings AttributeLayerSettings
	// EnvironmentAttributes is the list of environment attributes in this layer.
	EnvironmentAttributes []EnvironmentAttributeData
}

// Marshal encodes/decodes an AttributeLayerData.
func (x *AttributeLayerData) Marshal(r IO) {
	r.String(&x.Name)
	r.Varint32(&x.DimensionID)
	Single(r, &x.Settings)
	Slice(r, &x.EnvironmentAttributes)
}
