package protocol

// TrimPattern represents a pattern that can be applied to an armour piece in combination with a TrimMaterial.
type TrimPattern struct {
	// ItemName is the identifier of the item that represents the pattern, for example
	// 'minecraft:wayfinder_armor_trim_smithing_template'.
	ItemName string
	// PatternID is the identifier of the pattern, for example, 'wayfinder'.
	PatternID string
}

// Marshal ...
func (x *TrimPattern) Marshal(r IO) {
	r.String(&x.ItemName)
	r.String(&x.PatternID)
}

// TrimMaterial represents a material that can be used when applying an armour trim.
type TrimMaterial struct {
	// MaterialID is the identifier of the material, for example 'netherite'.
	MaterialID string
	// Colour is the colour code used for text formatting, for example '§j'.
	Colour string
	// ItemName is the identifier of the item that represents the material, for example, 'minecraft:netherite_ingot'.
	ItemName string
}

// Marshal ...
func (x *TrimMaterial) Marshal(r IO) {
	r.String(&x.MaterialID)
	r.String(&x.Colour)
	r.String(&x.ItemName)
}
