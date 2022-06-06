package protocol

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
	// Range is the height range of the dimension, where the first value is the minimum and the second is the maximum.
	Range [2]int32
	// Generator is the variant of generator that exists in the provided dimension. These can be one of the constants
	// defined above. If this is set to GeneratorLegacy, the legacy horizontal world limits will be enforced.
	Generator int32
}

// DimensionDef reads/writes a DimensionDefinition x using IO r.
func DimensionDef(r IO, x *DimensionDefinition) {
	r.String(&x.Name)
	r.Varint32(&x.Range[0])
	r.Varint32(&x.Range[1])
	r.Varint32(&x.Generator)
}
