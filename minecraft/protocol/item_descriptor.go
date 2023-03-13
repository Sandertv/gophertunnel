package protocol

// ItemDescriptorCount represents an item descriptor that has a count attached with it, such as a recipe ingredient.
type ItemDescriptorCount struct {
	// Descriptor represents how the item is described over the network. It is one of the descriptors above.
	Descriptor ItemDescriptor
	// Count is the count of items that the item descriptor is required to have.
	Count int32
}

// ItemDescriptor represents a type of item descriptor. This is one of the concrete types below. It is an alias of
// Marshaler.
type ItemDescriptor interface {
	Marshaler
}

const (
	ItemDescriptorInvalid = iota
	ItemDescriptorDefault
	ItemDescriptorMoLang
	ItemDescriptorItemTag
	ItemDescriptorDeferred
	ItemDescriptorComplexAlias
)

// InvalidItemDescriptor represents an invalid item descriptor. This is usually sent by the vanilla server for empty
// slots or ingredients.
type InvalidItemDescriptor struct{}

// Marshal ...
func (*InvalidItemDescriptor) Marshal(IO) {}

// DefaultItemDescriptor represents an item descriptor for regular items. This is used for the significant majority of
// items.
type DefaultItemDescriptor struct {
	// NetworkID is the numerical network ID of the item. This is sometimes a positive ID, and sometimes a
	// negative ID, depending on what item it concerns.
	NetworkID int16
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int16
}

// Marshal ...
func (x *DefaultItemDescriptor) Marshal(r IO) {
	r.Int16(&x.NetworkID)
	if x.NetworkID != 0 {
		r.Int16(&x.MetadataValue)
	}
}

// MoLangItemDescriptor represents an item descriptor for items that use MoLang (e.g. behaviour packs).
type MoLangItemDescriptor struct {
	// Expression represents the MoLang expression used to identify the item/it's associated tag.
	Expression string
	// Version represents the version of MoLang to use.
	Version uint8
}

// Marshal ...
func (x *MoLangItemDescriptor) Marshal(r IO) {
	r.String(&x.Expression)
	r.Uint8(&x.Version)
}

// ItemTagItemDescriptor represents an item descriptor that uses item tagging. This should be used to reduce duplicative
// entries for items that can be grouped under a single tag.
type ItemTagItemDescriptor struct {
	// Tag represents the tag that the item is part of.
	Tag string
}

// Marshal ...
func (x *ItemTagItemDescriptor) Marshal(r IO) {
	r.String(&x.Tag)
}

// DeferredItemDescriptor represents an item descriptor that uses a namespace and metadata value to identify the item.
// There is no clear benefit of using this item descriptor.
type DeferredItemDescriptor struct {
	// Name is the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int16
}

// Marshal ...
func (x *DeferredItemDescriptor) Marshal(r IO) {
	r.String(&x.Name)
	r.Int16(&x.MetadataValue)
}

// ComplexAliasItemDescriptor represents an item descriptor that uses a single name to identify the item. There is no
// clear benefit of using this item descriptor and only seem to be used for specific recipes.
type ComplexAliasItemDescriptor struct {
	// Name is the name of the item, which is a name like 'minecraft:stick'.
	Name string
}

// Marshal ...
func (x *ComplexAliasItemDescriptor) Marshal(r IO) {
	r.String(&x.Name)
}
