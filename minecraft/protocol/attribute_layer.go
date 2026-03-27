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

const (
	AttributeLayerWeightTypeFloat = iota
	AttributeLayerWeightTypeString
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

// EnvironmentAttributeData represents an environment attribute with optional transition data.
type EnvironmentAttributeData struct {
	// AttributeName is the name of the attribute.
	AttributeName string
	// FromAttribute is the optional starting attribute for transitions.
	FromAttribute Optional[AttributeData]
	// Attribute is the current attribute value.
	Attribute AttributeData
	// ToAttribute is the optional target attribute for transitions.
	ToAttribute Optional[AttributeData]
	// CurrentTransitionTicks is the number of ticks elapsed in the current transition.
	CurrentTransitionTicks uint32
	// TotalTransitionTicks is the total number of ticks for the transition.
	TotalTransitionTicks uint32
	// EaseType is the easing function used for the transition. It is one of the EasingType constants.
	EaseType int32
}

// Marshal encodes/decodes an EnvironmentAttributeData.
func (x *EnvironmentAttributeData) Marshal(r IO) {
	easingType := easingTypeToString(x.EaseType)
	r.String(&x.AttributeName)
	OptionalMarshaler(r, &x.FromAttribute)
	Single(r, &x.Attribute)
	OptionalMarshaler(r, &x.ToAttribute)
	r.Uint32(&x.CurrentTransitionTicks)
	r.Uint32(&x.TotalTransitionTicks)
	r.String(&easingType)
	easingTypeFromString(r, &x.EaseType, easingType)
}

// AttributeLayerSettings represents settings for an attribute layer.
type AttributeLayerSettings struct {
	// Priority is the priority of the layer.
	Priority int32
	// WeightType determines whether the weight is a float or string. It is one of the
	// AttributeLayerWeightType constants.
	WeightType uint32
	// FloatWeight is the weight if WeightType is AttributeLayerWeightTypeFloat.
	FloatWeight float32
	// StringWeight is the weight if WeightType is AttributeLayerWeightTypeString.
	StringWeight string
	// Enabled indicates if the layer is enabled.
	Enabled bool
	// TransitionsPaused indicates if transitions are paused for this layer.
	TransitionsPaused bool
}

// Marshal encodes/decodes an AttributeLayerSettings.
func (x *AttributeLayerSettings) Marshal(r IO) {
	r.Int32(&x.Priority)
	r.Varuint32(&x.WeightType)
	switch x.WeightType {
	case AttributeLayerWeightTypeFloat:
		r.Float32(&x.FloatWeight)
	case AttributeLayerWeightTypeString:
		r.String(&x.StringWeight)
	default:
		r.UnknownEnumOption(x.WeightType, "attribute layer weight type")
	}
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
