package protocol

import "fmt"

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
)

// InvalidItemDescriptor represents an invalid item descriptor. This is usually sent by the vanilla server for empty
// slots or ingredients.
type InvalidItemDescriptor struct{}

// Marshal ...
func (*InvalidItemDescriptor) Marshal(IO) {}

// DefaultItemDescriptor represents an item descriptor for regular items. This is used for the significant majority of
// items.
type DefaultItemDescriptor struct {
	// Name is the identifier of the item, such as minecraft:stone.
	Name string
	// MetadataValue is the metadata value of the item. For some items, this is the damage value, whereas for
	// other items it is simply an identifier of a variant of the item.
	MetadataValue int32
}

// Marshal ...
func (x *DefaultItemDescriptor) Marshal(r IO) {
	r.String(&x.Name)
	r.Varint32(&x.MetadataValue)
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

func itemDescriptorType(x ItemDescriptor) (uint8, string, bool) {
	switch x.(type) {
	case nil, *InvalidItemDescriptor:
		return ItemDescriptorInvalid, "empty", true
	case *DefaultItemDescriptor:
		return ItemDescriptorDefault, "name", true
	case *MoLangItemDescriptor:
		return ItemDescriptorMoLang, "molang", true
	case *ItemTagItemDescriptor:
		return ItemDescriptorItemTag, "item_tag", true
	default:
		return 0, "", false
	}
}

func itemDescriptorFromType(id uint8) (ItemDescriptor, bool) {
	switch id {
	case ItemDescriptorInvalid:
		return &InvalidItemDescriptor{}, true
	case ItemDescriptorDefault:
		return &DefaultItemDescriptor{}, true
	case ItemDescriptorMoLang:
		return &MoLangItemDescriptor{}, true
	case ItemDescriptorItemTag:
		return &ItemTagItemDescriptor{}, true
	default:
		return nil, false
	}
}

// StackRequestItemDescriptorCount reads/writes the Cereal item descriptor format used by auto-craft actions.
func StackRequestItemDescriptorCount(r IO, x *ItemDescriptorCount) {
	id, _, ok := itemDescriptorType(x.Descriptor)
	if !ok {
		r.UnknownEnumOption(fmt.Sprintf("%T", x.Descriptor), "stack request item descriptor type")
		return
	}
	variant := uint32(id)
	r.Varuint32(&variant)
	legacyID := id
	r.Uint8(&legacyID)
	if variant > ItemDescriptorItemTag {
		r.UnknownEnumOption(variant, "stack request item descriptor type")
		return
	}
	descriptor := x.Descriptor
	if descriptor == nil {
		descriptor = &InvalidItemDescriptor{}
	}
	if uint32(id) != variant {
		var found bool
		descriptor, found = itemDescriptorFromType(uint8(variant))
		if !found {
			r.UnknownEnumOption(variant, "stack request item descriptor type")
			return
		}
		x.Descriptor = descriptor
	}
	descriptor.Marshal(r)
	IntegerFunc(&x.Count, r.Uint16)
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
