package protocol

// EnchantmentOption represents a single option in the enchantment table for a single item.
type EnchantmentOption struct {
	// Cost is the cost of the option. This is the amount of XP levels required to select this enchantment
	// option.
	Cost uint32
	// Enchantments holds the enchantments that will be applied to the item when this option is clicked.
	Enchantments ItemEnchantments
	// Name is a name that will be translated to the 'Standard Galactic Alphabet'
	// (https://minecraft.gamepedia.com/Enchanting_Table#Standard_Galactic_Alphabet) client-side. The names
	// generally have no meaning, such as:
	// 'animal imbue range galvanize '
	// 'bless inside creature shrink '
	// 'elder free of inside '
	Name string
	// RecipeNetworkID is a unique network ID for this enchantment option. When enchanting, the client
	// will submit this network ID in a ItemStackRequest packet with the CraftRecipe action, so that the
	// server knows which enchantment was selected.
	// Note that this ID should still be unique with other actual recipes. It's recommended to start counting
	// for enchantment network IDs from the counter used for producing network IDs for the normal recipes.
	RecipeNetworkID uint32
}

// WriteEnchantOption writes an EnchantmentOption x to Writer w.
func WriteEnchantOption(w *Writer, x *EnchantmentOption) {
	w.Varuint32(&x.Cost)
	WriteItemEnchants(w, &x.Enchantments)
	w.String(&x.Name)
	w.Varuint32(&x.RecipeNetworkID)
}

// EnchantOption reads an EnchantmentOption x from Reader r.
func EnchantOption(r *Reader, x *EnchantmentOption) {
	r.Varuint32(&x.Cost)
	ItemEnchants(r, &x.Enchantments)
	r.String(&x.Name)
	r.Varuint32(&x.RecipeNetworkID)
}

// ItemEnchantments holds information on the enchantments that are applied to an item when a specific button
// is clicked in the enchantment table.
type ItemEnchantments struct {
	// Slot is either 0, 1 or 2. Its exact usage is not clear.
	Slot int32
	// Enchantments is an array of 3 slices of enchantment instances. Each array represents enchantments that
	// will be added to the item with a different activation type. The arrays in which enchantments are sent
	// by the vanilla server are as follows:
	// slice 1 { protection, fire protection, feather falling, blast protection, projectile protection,
	//           thorns, respiration, depth strider, aqua affinity, frost walker, soul speed }
	// slice 2 { sharpness, smite, bane of arthropods, fire aspect, looting, silk touch, unbreaking, fortune,
	//           flame, luck of the sea, impaling }
	// slice 3 { knockback, efficiency, power, punch, infinity, lure, mending, curse of binding,
	//           curse of vanishing, riptide, loyalty, channeling, multishot, piercing, quick charge }
	// The first slice holds armour enchantments, the differences between the slice 2 and slice 3 are more
	// vaguely defined.
	Enchantments [3][]EnchantmentInstance
}

// WriteItemEnchants writes an ItemEnchantments x to Writer w..
func WriteItemEnchants(w *Writer, x *ItemEnchantments) {
	w.Int32(&x.Slot)
	for _, enchantments := range x.Enchantments {
		l := uint32(len(enchantments))
		w.Varuint32(&l)
		for _, enchantment := range enchantments {
			Enchant(w, &enchantment)
		}
	}
}

// ItemEnchants reads an ItemEnchantments x from Reader r.
func ItemEnchants(r *Reader, x *ItemEnchantments) {
	var l uint32
	r.Int32(&x.Slot)
	for i := 0; i < 3; i++ {
		r.Varuint32(&l)
		x.Enchantments[i] = make([]EnchantmentInstance, l)
		for j := uint32(0); j < l; j++ {
			Enchant(r, &x.Enchantments[i][j])
		}
	}
}

// EnchantmentInstance represents a single enchantment instance with the type of the enchantment and its
// level.
type EnchantmentInstance struct {
	Type  byte
	Level byte
}

// Enchant reads/writes an EnchantmentInstance x using IO r.
func Enchant(r IO, x *EnchantmentInstance) {
	r.Uint8(&x.Type)
	r.Uint8(&x.Level)
}
