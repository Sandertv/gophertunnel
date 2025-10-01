package protocol

const (
	BiomeExpressionOpUnknown = iota - 1
	BiomeExpressionOpLeftBrace
	BiomeExpressionOpRightBrace
	BiomeExpressionOpLeftBracket
	BiomeExpressionOpRightBracket
	BiomeExpressionOpLeftParenthesis
	BiomeExpressionOpRightParenthesis
	BiomeExpressionOpNegate
	BiomeExpressionOpLogicalNot
	BiomeExpressionOpAbs
	BiomeExpressionOpAdd
	BiomeExpressionOpAcos
	BiomeExpressionOpAsin
	BiomeExpressionOpAtan
	BiomeExpressionOpAtan2
	BiomeExpressionOpCeil
	BiomeExpressionOpClamp
	BiomeExpressionOpCopySign
	BiomeExpressionCos
	BiomeExpressionDieRoll
	BiomeExpressionDieRollInt
	BiomeExpressionDiv
	BiomeExpressionExp
	BiomeExpressionFloor
	BiomeExpressionHermiteBlend
	BiomeExpressionLerp
	BiomeExpressionLerpRotate
	BiomeExpressionLn
	BiomeExpressionMax
	BiomeExpressionMin
	BiomeExpressionMinAngle
	BiomeExpressionMod
	BiomeExpressionMul
	BiomeExpressionPow
	BiomeExpressionRandom
	BiomeExpressionRandomInt
	BiomeExpressionRound
	BiomeExpressionSin
	BiomeExpressionSign
	BiomeExpressionSqrt
	BiomeExpressionTrunc
	BiomeExpressionQueryFunction
	BiomeExpressionArrayVariable
	BiomeExpressionContextVariable
	BiomeExpressionEntityVariable
	BiomeExpressionTempVariable
	BiomeExpressionMemberAccessor
	BiomeExpressionHashedStringHash
	BiomeExpressionGeometryVariable
	BiomeExpressionMaterialVariable
	BiomeExpressionTextureVariable
	BiomeExpressionLessThan
	BiomeExpressionLessEqual
	BiomeExpressionGreaterEqual
	BiomeExpressionGreaterThan
	BiomeExpressionLogicalEqual
	BiomeExpressionLogicalNotEqual
	BiomeExpressionLogicalOr
	BiomeExpressionLogicalAnd
	BiomeExpressionNullCoalescing
	BiomeExpressionConditional
	BiomeExpressionConditionalElse
	BiomeExpressionFloat
	BiomeExpressionPi
	BiomeExpressionArray
	BiomeExpressionGeometry
	BiomeExpressionMaterial
	BiomeExpressionTexture
	BiomeExpressionLoop
	BiomeExpressionForEach
	BiomeExpressionBreak
	BiomeExpressionContinue
	BiomeExpressionAssignment
	BiomeExpressionPointer
	BiomeExpressionSemicolon
	BiomeExpressionReturn
	BiomeExpressionComma
	BiomeExpressionThis
)

const (
	BiomeCoordinateEvaluationOrderXYZ = iota
	BiomeCoordinateEvaluationOrderXZY
	BiomeCoordinateEvaluationOrderYXZ
	BiomeCoordinateEvaluationOrderYZX
	BiomeCoordinateEvaluationOrderZXY
	BiomeCoordinateEvaluationOrderZYX
)

const (
	BiomeRandomDistributionTypeSingleValued = iota
	BiomeRandomDistributionTypeUniform
	BiomeRandomDistributionTypeGaussian
	BiomeRandomDistributionTypeInverseGaussian
	BiomeRandomDistributionTypeFixedGrid
	BiomeRandomDistributionTypeJitteredGrid
	BiomeRandomDistributionTypeTriangle
)

// BiomeDefinition represents a biome definition in the game. This can be a vanilla biome or a completely
// custom biome.
type BiomeDefinition struct {
	// NameIndex represents the index of the biome name in the string list.
	NameIndex int16
	// BiomeID is the biome ID.
	BiomeID int16
	// Temperature is the temperature of the biome, used for weather, biome behaviours and sky colour.
	Temperature float32
	// Downfall is the amount that precipitation affects colours and block changes.
	Downfall float32
	// FoliageSnow is the progression factor for foliage turning white due to snow.
	FoliageSnow float32
	// Depth is the depth of the biome.
	Depth float32
	// Scale is the scale of the biome.
	Scale float32
	// MapWaterColour is an ARGB value for the water colour on maps in the biome.
	MapWaterColour int32
	// Rain is true if the biome has rain, false if it is a dry biome.
	Rain bool
	// Tags are a list of indices of tags in the string list. These are used to group biomes together for
	// biome generation and other purposes.
	Tags Optional[[]uint16]
	// ChunkGeneration is optional information to assist in client-side chunk generation. Almost all servers
	// can and should leave this empty to greatly reduce the size of this packet. Only BDS and servers which
	// *exactly* match the vanilla chunk generation can benefit from this.
	ChunkGeneration Optional[BiomeChunkGeneration]
}

func (x *BiomeDefinition) Marshal(r IO) {
	r.Int16(&x.NameIndex)
	r.Int16(&x.BiomeID)
	r.Float32(&x.Temperature)
	r.Float32(&x.Downfall)
	r.Float32(&x.FoliageSnow)
	r.Float32(&x.Depth)
	r.Float32(&x.Scale)
	r.Int32(&x.MapWaterColour)
	r.Bool(&x.Rain)
	OptionalFunc(r, &x.Tags, func(s *[]uint16) {
		FuncSlice(r, s, r.Uint16)
	})
	OptionalMarshaler(r, &x.ChunkGeneration)
}

// BiomeChunkGeneration represents the information required for the client to generate chunks itself
// to create the illusion of a larger render distance.
type BiomeChunkGeneration struct {
	// Climate is optional information to specify the biome's climate.
	Climate Optional[BiomeClimate]
	// ConsolidatedFeatures is a list of features that are consolidated into a single feature.
	ConsolidatedFeatures Optional[[]BiomeConsolidatedFeature]
	// MountainParameters is optional information to specify the biome's mountain parameters.
	MountainParameters Optional[BiomeMountainParameters]
	// SurfaceMaterialAdjustments is a list of surface material adjustments.
	SurfaceMaterialAdjustments Optional[[]BiomeElementData]
	// SurfaceMaterials is a set of materials to use for the surface layers of the biome.
	SurfaceMaterials Optional[BiomeSurfaceMaterial]
	// HasDefaultOverworldSurface is true if the biome has a default overworld surface.
	HasDefaultOverworldSurface bool
	// HasSwampSurface is true if the biome has a swamp surface.
	HasSwampSurface bool
	// HasFrozenOceanSurface is true if the biome has a frozen ocean surface.
	HasFrozenOceanSurface bool
	// HasEndSurface is true if the biome has an end surface.
	HasEndSurface bool
	// MesaSurface is optional information to specify the biome's mesa surface.
	MesaSurface Optional[BiomeMesaSurface]
	// CappedSurface is optional information to specify the biome's capped surface, i.e. in the Nether.
	CappedSurface Optional[BiomeCappedSurface]
	// OverworldRules is optional information to specify the biome's overworld rules, such as rivers and hills.
	OverworldRules Optional[BiomeOverworldRules]
	// MultiNoiseRules is optional information to specify the biome's multi-noise rules.
	MultiNoiseRules Optional[BiomeMultiNoiseRules]
	// LegacyRules is a list of legacy rules for the biomes using an older format, which is just a list of
	// weighted biomes.
	LegacyRules Optional[[]BiomeConditionalTransformation]
}

func (x *BiomeChunkGeneration) Marshal(r IO) {
	OptionalMarshaler(r, &x.Climate)
	OptionalFunc(r, &x.ConsolidatedFeatures, func(s *[]BiomeConsolidatedFeature) {
		Slice(r, s)
	})
	OptionalMarshaler(r, &x.MountainParameters)
	OptionalFunc(r, &x.SurfaceMaterialAdjustments, func(s *[]BiomeElementData) {
		Slice(r, s)
	})
	OptionalMarshaler(r, &x.SurfaceMaterials)
	r.Bool(&x.HasDefaultOverworldSurface)
	r.Bool(&x.HasSwampSurface)
	r.Bool(&x.HasFrozenOceanSurface)
	r.Bool(&x.HasEndSurface)
	OptionalMarshaler(r, &x.MesaSurface)
	OptionalMarshaler(r, &x.CappedSurface)
	OptionalMarshaler(r, &x.OverworldRules)
	OptionalMarshaler(r, &x.MultiNoiseRules)
	OptionalFunc(r, &x.LegacyRules, func(s *[]BiomeConditionalTransformation) {
		Slice(r, s)
	})
}

// BiomeClimate represents the climate of a biome, mainly for ambience but also defines certain behaviours.
type BiomeClimate struct {
	// Temperature is the temperature of the biome, used for weather, biome behaviours and sky colour.
	Temperature float32
	// Downfall is the amount that precipitation affects colours and block changes.
	Downfall float32
	// SnowAccumulationMin is the minimum amount of snow that can accumulate in the biome, every 0.125 is
	// another layer of snow.
	SnowAccumulationMin float32
	// SnowAccumulationMax is the maximum amount of snow that can accumulate in the biome, every 0.125 is
	// another layer of snow.
	SnowAccumulationMax float32
}

func (x *BiomeClimate) Marshal(r IO) {
	r.Float32(&x.Temperature)
	r.Float32(&x.Downfall)
	r.Float32(&x.SnowAccumulationMin)
	r.Float32(&x.SnowAccumulationMax)
}

// BiomeConsolidatedFeature represents a feature that is consolidated into a single feature for the biome.
type BiomeConsolidatedFeature struct {
	// Scatter defines how the feature is scattered in the biome.
	Scatter BiomeScatterParameter
	// Feature is the index of the feature's name in the string list.
	Feature int16
	// Identifier is the index of the feature's identifier in the string list.
	Identifier int16
	// Pass is the index of the feature's pass in the string list.
	Pass int16
	// CanUseInternal is true if the feature can use internal features.
	CanUseInternal bool
}

func (x *BiomeConsolidatedFeature) Marshal(r IO) {
	Single(r, &x.Scatter)
	r.Int16(&x.Feature)
	r.Int16(&x.Identifier)
	r.Int16(&x.Pass)
	r.Bool(&x.CanUseInternal)
}

type BiomeScatterParameter struct {
	// Coordinates is a list of coordinate rules to scatter the feature within.
	Coordinates []BiomeCoordinate
	// EvaluationOrder is the order in which the coordinates are evaluated, and is one of the
	// CoordinateEvaluationOrder constants above.
	EvaluationOrder int32
	// ChancePercentType is the type of expression operation to use for the chance percent, and is one of the
	// BiomeExpressionOp constants above.
	ChancePercentType int32
	// ChangePercent is the index of the chance expression in the string list.
	ChancePercent int16
	// ChanceNumerator is the numerator of the chance expression.
	ChanceNumerator int32
	// ChanceDenominator is the denominator of the chance expression.
	ChanceDenominator int32
	// IterationsType is the type of expression operation to use for the iterations, and is one of the
	// BiomeExpressionOp constants above.
	IterationsType int32
	// Iterations is the index of the iterations expression in the string list.
	Iterations int16
}

func (x *BiomeScatterParameter) Marshal(r IO) {
	Slice(r, &x.Coordinates)
	r.Varint32(&x.EvaluationOrder)
	r.Varint32(&x.ChancePercentType)
	r.Int16(&x.ChancePercent)
	r.Int32(&x.ChanceNumerator)
	r.Int32(&x.ChanceDenominator)
	r.Varint32(&x.IterationsType)
	r.Int16(&x.Iterations)
}

// BiomeCoordinate specifies coordinate rules for where features can be scattered in the biome.
type BiomeCoordinate struct {
	// MinValueType is the type of expression operation to use for the minimum value, and is one of the
	// BiomeExpressionOp constants above.
	MinValueType int32
	// MinValue is the index of the minimum value expression in the string list.
	MinValue int16
	// MaxValueType is the type of expression operation to use for the maximum value, and is one of the
	MaxValueType int32
	// MaxValue is the index of the maximum value expression in the string list.
	MaxValue int16
	// GridOffset is the offset of the grid, used for fixed grid and jittered grid distributions.
	GridOffset uint32
	// GridStepSize is the step size of the grid, used for fixed grid and jittered grid distributions.
	GridStepSize uint32
	// Distribution is the type of distribution to use for the coordinate, and is one of the
	// BiomeRandomDistributionType constants above.
	Distribution int32
}

func (x *BiomeCoordinate) Marshal(r IO) {
	r.Varint32(&x.MinValueType)
	r.Int16(&x.MinValue)
	r.Varint32(&x.MaxValueType)
	r.Int16(&x.MaxValue)
	r.Uint32(&x.GridOffset)
	r.Uint32(&x.GridStepSize)
	r.Varint32(&x.Distribution)
}

// BiomeMountainParameters specifies the parameters for a mountain biome.
type BiomeMountainParameters struct {
	// SteepBlock is the runtime ID of the block to use for steep slopes.
	SteepBlock int32
	// NorthSlopes is true if the biome has north slopes.
	NorthSlopes bool
	// SouthSlopes is true if the biome has south slopes.
	SouthSlopes bool
	// WestSlopes is true if the biome has west slopes.
	WestSlopes bool
	// EastSlopes is true if the biome has east slopes.
	EastSlopes bool
	// TopSlideEnabled is true if the biome has top slide enabled.
	TopSlideEnabled bool
}

func (x *BiomeMountainParameters) Marshal(r IO) {
	r.Int32(&x.SteepBlock)
	r.Bool(&x.NorthSlopes)
	r.Bool(&x.SouthSlopes)
	r.Bool(&x.WestSlopes)
	r.Bool(&x.EastSlopes)
	r.Bool(&x.TopSlideEnabled)
}

// BiomeElementData are set rules to adjust the surface materials of the biome.
type BiomeElementData struct {
	// NoiseFrequencyScale is the frequency scale of the noise used to adjust the surface materials.
	NoiseFrequencyScale float32
	// NoiseLowerBound is the minimum noise value required to be selected.
	NoiseLowerBound float32
	// NoiseUpperBound is the maximum noise value required to be selected.
	NoiseUpperBound float32
	// HeightMinType is the type of expression operation to use for the minimum height, and is one of the
	// BiomeExpressionOp constants above.
	HeightMinType int32
	// HeightMin is the index of the minimum height expression in the string list.
	HeightMin int16
	// HeightMaxType is the type of expression operation to use for the maximum height, and is one of the
	// BiomeExpressionOp constants above.
	HeightMaxType int32
	// HeightMax is the index of the maximum height expression in the string list.
	HeightMax int16
	// AdjustedMaterials is the materials to use for the surface layers of the biome if selected.
	AdjustedMaterials BiomeSurfaceMaterial
}

func (x *BiomeElementData) Marshal(r IO) {
	r.Float32(&x.NoiseFrequencyScale)
	r.Float32(&x.NoiseLowerBound)
	r.Float32(&x.NoiseUpperBound)
	r.Varint32(&x.HeightMinType)
	r.Int16(&x.HeightMin)
	r.Varint32(&x.HeightMaxType)
	r.Int16(&x.HeightMax)
	Single(r, &x.AdjustedMaterials)
}

// BiomeSurfaceMaterial specifies the materials to use for the surface layers of the biome.
type BiomeSurfaceMaterial struct {
	// TopBlock is the runtime ID of the block to use for the top layer.
	TopBlock int32
	// MidBlock is the runtime ID to use for the middle layers.
	MidBlock int32
	// SeaFloorBlock is the runtime ID to use for the sea floor.
	SeaFloorBlock int32
	// FoundationBlock is the runtime ID to use for the foundation layers.
	FoundationBlock int32
	// SeaBlock is the runtime ID to use for the sea layers.
	SeaBlock int32
	// SeaFloorDepth is the depth of the sea floor, in blocks.
	SeaFloorDepth int32
}

func (x *BiomeSurfaceMaterial) Marshal(r IO) {
	r.Int32(&x.TopBlock)
	r.Int32(&x.MidBlock)
	r.Int32(&x.SeaFloorBlock)
	r.Int32(&x.FoundationBlock)
	r.Int32(&x.SeaBlock)
	r.Int32(&x.SeaFloorDepth)
}

// BiomeMesaSurface specifies the materials to use for the mesa biome.
type BiomeMesaSurface struct {
	// ClayMaterial is the runtime ID of the block to use for clay layers.
	ClayMaterial uint32
	// HardClayMaterial is the runtime ID of the block to use for hard clay layers.
	HardClayMaterial uint32
	// BrycePillars is true if the biome has bryce pillars, which are tall spire-like structures.
	BrycePillars bool
	// HasForest is true if the biome has a forest.
	HasForest bool
}

func (x *BiomeMesaSurface) Marshal(r IO) {
	r.Uint32(&x.ClayMaterial)
	r.Uint32(&x.HardClayMaterial)
	r.Bool(&x.BrycePillars)
	r.Bool(&x.HasForest)
}

// BiomeCappedSurface specifies the materials to use for the capped surface of a biome, such as in the Nether.
type BiomeCappedSurface struct {
	// FloorBlocks is a list of runtime IDs to use for the floor blocks.
	FloorBlocks []int32
	// CeilingBlocks is a list of runtime IDs to use for the ceiling blocks.
	CeilingBlocks []int32
	// SeaBlock is an optional runtime ID to use for the sea block.
	SeaBlock Optional[uint32]
	// FoundationBlock is an optional runtime ID to use for the foundation block.
	FoundationBlock Optional[uint32]
	// BeachBlock is an optional runtime ID to use for the beach block.
	BeachBlock Optional[uint32]
}

func (x *BiomeCappedSurface) Marshal(r IO) {
	FuncSlice(r, &x.FloorBlocks, r.Int32)
	FuncSlice(r, &x.CeilingBlocks, r.Int32)
	OptionalFunc(r, &x.SeaBlock, r.Uint32)
	OptionalFunc(r, &x.FoundationBlock, r.Uint32)
	OptionalFunc(r, &x.BeachBlock, r.Uint32)
}

// BiomeOverworldRules specifies a list of transformation rules to apply to different parts of the overworld.
type BiomeOverworldRules struct {
	// HillsTransformations is a list of weighted biome transformations to apply to hills.
	HillsTransformations []BiomeWeight
	// MutateTransformations is a list of weighted biome transformations to apply to mutated biomes.
	MutateTransformations []BiomeWeight
	// RiverTransformations is a list of weighted biome transformations to apply to rivers.
	RiverTransformations []BiomeWeight
	// ShoreTransformations is a list of weighted biome transformations to apply to shores.
	ShoreTransformations []BiomeWeight
	// PreHillsEdgeTransformations is a list of conditional transformations to apply to the edges of hills.
	PreHillsEdgeTransformations []BiomeConditionalTransformation
	// PostShoreEdgeTransformations is a list of conditional transformations to apply to the edges of shores.
	PostShoreEdgeTransformations []BiomeConditionalTransformation
	// ClimateTransformations is a list of weighted temperature transformations to apply to the biome's climate.
	ClimateTransformations []BiomeTemperatureWeight
}

func (x *BiomeOverworldRules) Marshal(r IO) {
	Slice(r, &x.HillsTransformations)
	Slice(r, &x.MutateTransformations)
	Slice(r, &x.RiverTransformations)
	Slice(r, &x.ShoreTransformations)
	Slice(r, &x.PreHillsEdgeTransformations)
	Slice(r, &x.PostShoreEdgeTransformations)
	Slice(r, &x.ClimateTransformations)
}

// BiomeMultiNoiseRules specifies the rules for multi-noise biomes, which are biomes that are defined by
// multiple noise parameters instead of just temperature and humidity.
type BiomeMultiNoiseRules struct {
	// Temperature is the temperature level of the biome.
	Temperature float32
	// Humidity is the humidity level of the biome.
	Humidity float32
	// Altitude is the altitude level of the biome.
	Altitude float32
	// Weirdness is the weirdness level of the biome.
	Weirdness float32
	// Weight is the weight of the biome, with a higher weight being more likely to be selected.
	Weight float32
}

func (x *BiomeMultiNoiseRules) Marshal(r IO) {
	r.Float32(&x.Temperature)
	r.Float32(&x.Humidity)
	r.Float32(&x.Altitude)
	r.Float32(&x.Weirdness)
	r.Float32(&x.Weight)
}

// BiomeConditionalTransformation is the legacy method of transforming biomes.
type BiomeConditionalTransformation struct {
	// WeightedBiomes is a list of biomes and their weights.
	WeightedBiomes []BiomeWeight
	// ConditionJSON is an index of the condition JSON data in the string list.
	ConditionJSON int16
	// MinPassingNeighbours is the minimum number of neighbours that must pass the condition for the
	// transformation to be applied.
	MinPassingNeighbours uint32
}

func (x *BiomeConditionalTransformation) Marshal(r IO) {
	Slice(r, &x.WeightedBiomes)
	r.Int16(&x.ConditionJSON)
	r.Uint32(&x.MinPassingNeighbours)
}

// BiomeWeight defines the weight for a biome, used for weighted randomness.
type BiomeWeight struct {
	// Biome is the index of the biome name in the string list.
	Biome int16
	// Weight is the weight of the biome, with a higher weight being more likely to be selected.
	Weight uint32
}

func (x *BiomeWeight) Marshal(r IO) {
	r.Int16(&x.Biome)
	r.Uint32(&x.Weight)
}

// BiomeTemperatureWeight defines the weight for a temperature, used for weighted randomness.
type BiomeTemperatureWeight struct {
	// Temperature is the temperature that can be selected.
	Temperature int32
	// Weight is the weight of the temperature, with a higher weight being more likely to be selected.
	Weight uint32
}

func (x *BiomeTemperatureWeight) Marshal(r IO) {
	r.Varint32(&x.Temperature)
	r.Uint32(&x.Weight)
}
