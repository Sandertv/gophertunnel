package protocol

import "github.com/google/uuid"

const (
	GeneratorLegacy    = 0
	GeneratorOverworld = 1
	GeneratorFlat      = 2
	GeneratorNether    = 3
	GeneratorEnd       = 4
	GeneratorVoid      = 5
)

// DimensionDefinition contains information specifying dimension-specific properties, used for data-driven dimensions.
// These include the range (the height min/max), generator variant, and more.
type DimensionDefinition struct {
	// Name specifies the name of the dimension.
	Name string
	// MinimumY is the lowest Y coordinate that exists in the dimension.
	MinimumY int32
	// HeightRange is the number of blocks above MinimumY that exist in the dimension, so that the highest Y
	// coordinate in the dimension is MinimumY + HeightRange.
	HeightRange int32
	// Generator is the variant of generator that exists in the provided dimension. These can be one of the constants
	// defined above. If this is set to GeneratorLegacy, the legacy horizontal world limits will be enforced.
	Generator int32
	// DimensionType is the numeric identifier of the dimension. This cannot override a vanilla dimension (0-2), but
	// custom dimensions should start from 1000 like vanilla.
	DimensionType int32
	// PackID is the UUID of the behaviour pack which has added the dimension.
	PackID uuid.UUID
	// DefaultBiome is the identifier of the biome that the dimension defaults to.
	DefaultBiome string
}

// Marshal encodes/decodes a DimensionDefinition.
func (x *DimensionDefinition) Marshal(r IO) {
	r.String(&x.Name)
	r.Varint32(&x.MinimumY)
	r.Varint32(&x.HeightRange)
	r.Varint32(&x.Generator)
	r.Varint32(&x.DimensionType)
	r.UUID(&x.PackID)
	r.String(&x.DefaultBiome)
}

// GenerationFeature represents a world generation feature, used when encoding the FeatureRegistry to the client.
type GenerationFeature struct {
	// Name is the name of the feature.
	Name string
	// JSON is the encoded JSON data instructing the client on how to generate the feature.
	JSON []byte
}

// Marshal encodes/decodes a GenerationFeature.
func (x *GenerationFeature) Marshal(r IO) {
	r.String(&x.Name)
	r.ByteSlice(&x.JSON)
}
