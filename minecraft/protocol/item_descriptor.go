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

// The names below identify the type of an item descriptor on the wire. An item
// descriptor that is not empty is preceded by one of them.
const (
	ItemDescriptorNameItemName = "name"
	ItemDescriptorNameMoLang   = "molang"
	ItemDescriptorNameItemTag  = "item_tag"
)

// The constants below identify the type of an item descriptor inside an ItemStackRequest, which selects the
// variant by index rather than by one of the names above.
const (
	ItemDescriptorTypeInvalid uint8 = iota
	ItemDescriptorTypeItemName
	ItemDescriptorTypeMoLang
	ItemDescriptorTypeItemTag
)

// ItemDescriptorAnyMetadata is the metadata value used by item descriptors that
// match any variant of an item.
const ItemDescriptorAnyMetadata = 32767

// lookupItemDescriptorType looks up the variant index of an ItemDescriptor, used by descriptors inside an
// ItemStackRequest.
func lookupItemDescriptorType(x ItemDescriptor, t *uint8) bool {
	switch x.(type) {
	case *InvalidItemDescriptor, nil:
		*t = ItemDescriptorTypeInvalid
	case *DeferredItemDescriptor:
		*t = ItemDescriptorTypeItemName
	case *MoLangItemDescriptor:
		*t = ItemDescriptorTypeMoLang
	case *ItemTagItemDescriptor:
		*t = ItemDescriptorTypeItemTag
	default:
		return false
	}
	return true
}

// lookupItemDescriptorName looks up the wire name of an ItemDescriptor. An empty
// name is returned for an InvalidItemDescriptor, which carries no name.
func lookupItemDescriptorName(x ItemDescriptor, name *string) bool {
	switch x.(type) {
	case *InvalidItemDescriptor, nil:
		*name = ""
	case *DeferredItemDescriptor:
		*name = ItemDescriptorNameItemName
	case *MoLangItemDescriptor:
		*name = ItemDescriptorNameMoLang
	case *ItemTagItemDescriptor:
		*name = ItemDescriptorNameItemTag
	default:
		return false
	}
	return true
}

// InvalidItemDescriptor represents an invalid item descriptor. This is usually sent by the vanilla server for empty
// slots or ingredients.
type InvalidItemDescriptor struct{}

// Marshal ...
func (*InvalidItemDescriptor) Marshal(r IO) {
	metadata := int32(ItemDescriptorAnyMetadata)
	r.Varint32(&metadata)
}

// MoLangItemDescriptor represents an item descriptor for items that use MoLang (e.g. behaviour packs).
type MoLangItemDescriptor struct {
	// Expression represents the MoLang expression used to identify the item/it's associated tag.
	Expression string
	// Version represents the version of MoLang to use.
	Version int16
}

// Marshal ...
func (x *MoLangItemDescriptor) Marshal(r IO) {
	r.String(&x.Expression)
	r.Int16(&x.Version)
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
	metadata := int32(ItemDescriptorAnyMetadata)
	r.Varint32(&metadata)
}

// DeferredItemDescriptor represents an item descriptor that uses a namespace and metadata value to identify the item.
type DeferredItemDescriptor struct {
	// Name is the name of the item, which is a name like 'minecraft:stick'.
	Name string
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item. ItemDescriptorAnyMetadata matches any
	// variant.
	MetadataValue int16
}

// Marshal ...
func (x *DeferredItemDescriptor) Marshal(r IO) {
	r.String(&x.Name)
	IntegerFunc(&x.MetadataValue, r.Varint32)
}
