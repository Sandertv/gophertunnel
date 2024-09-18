package protocol

const (
	AttributeModifierOperationAddition = iota
	AttributeModifierOperationMultiplyBase
	AttributeModifierOperationMultiplyTotal
	AttributeModifierOperationCap
)

const (
	AttributeModifierOperandMin = iota
	AttributeModifierOperandMax
	AttributeModifierOperandCurrent
)

// AttributeValue holds the value of an attribute, including the min and max.
type AttributeValue struct {
	// Name is the name of the attribute, for example 'minecraft:health'. These names must be identical to
	// the ones defined client-side.
	Name string
	// Value is the current value of the attribute. This value will be applied to the entity when sent in a
	// packet.
	Value float32
	// Max and Min specify the boundaries within the value of the attribute must be. The definition of these
	// fields differ per attribute. The maximum health of an entity may be changed, whereas the maximum
	// movement speed for example may not be.
	Max, Min float32
}

// Marshal encodes/decodes an AttributeValue.
func (x *AttributeValue) Marshal(r IO) {
	r.String(&x.Name)
	r.Float32(&x.Min)
	r.Float32(&x.Value)
	r.Float32(&x.Max)
}

// Attribute is an entity attribute, that holds specific data such as the health of the entity. Each attribute
// holds a default value, maximum and minimum value, name and its current value.
type Attribute struct {
	AttributeValue
	// DefaultMin is the default minimum value of the attribute. It's not clear why this field must be sent to
	// the client, but it is required regardless.
	DefaultMin float32
	// DefaultMax is the default maximum value of the attribute. It's not clear why this field must be sent to
	// the client, but it is required regardless.
	DefaultMax float32
	// Default is the default value of the attribute. It's not clear why this field must be sent to the
	// client, but it is required regardless.
	Default float32
	// Modifiers is a slice of AttributeModifiers that are applied to the attribute.
	Modifiers []AttributeModifier
}

// Marshal encodes/decodes an Attribute.
func (x *Attribute) Marshal(r IO) {
	r.Float32(&x.Min)
	r.Float32(&x.Max)
	r.Float32(&x.Value)
	r.Float32(&x.DefaultMin)
	r.Float32(&x.DefaultMax)
	r.Float32(&x.Default)
	r.String(&x.Name)
	Slice(r, &x.Modifiers)
}

// AttributeModifier temporarily buffs/debuffs a given attribute until the modifier is used. In vanilla, these are
// mainly used for effects.
type AttributeModifier struct {
	// ID is the unique ID of the modifier. It is used to identify the modifier in the packet.
	ID string
	// Name is the name of the attribute that is modified.
	Name string
	// Amount is the amount of difference between the current value of the attribute and the new value.
	Amount float32
	// Operation is the operation that is performed on the attribute. It can be addition, multiply base, multiply total
	// or cap.
	Operation int32
	// Operand ...
	// TODO: Figure out what this field is used for.
	Operand int32
	// Serializable ...
	// TODO: Figure out what this field is used for.
	Serializable bool
}

// Marshal encodes/decodes an AttributeModifier.
func (x *AttributeModifier) Marshal(r IO) {
	r.String(&x.ID)
	r.String(&x.Name)
	r.Float32(&x.Amount)
	r.Int32(&x.Operation)
	r.Int32(&x.Operand)
	r.Bool(&x.Serializable)
}
