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

// Attribute is an entity attribute, that holds specific data such as the health of the entity. Each attribute
// holds a default value, maximum and minimum value, name and its current value.
type Attribute struct {
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
	// Default is the default value of the attribute. It's not clear why this field must be sent to the
	// client, but it is required regardless.
	Default float32
	// Modifiers is a slice of AttributeModifiers that are applied to the attribute.
	Modifiers []AttributeModifier
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

// Attributes reads an Attribute slice x from Reader r.
func Attributes(r *Reader, x *[]Attribute) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)

	*x = make([]Attribute, count)
	for i := uint32(0); i < count; i++ {
		attribute := &(*x)[i]
		r.Float32(&attribute.Min)
		r.Float32(&attribute.Max)
		r.Float32(&attribute.Value)
		r.Float32(&attribute.Default)
		r.String(&attribute.Name)

		var modifierCount uint32
		r.Varuint32(&modifierCount)

		attribute.Modifiers = make([]AttributeModifier, modifierCount)
		for j := uint32(0); j < modifierCount; j++ {
			modifier := &(attribute.Modifiers)[j]
			r.String(&modifier.ID)
			r.String(&modifier.Name)
			r.Float32(&modifier.Amount)
			r.Int32(&modifier.Operation)
			r.Int32(&modifier.Operand)
			r.Bool(&modifier.Serializable)
		}
	}
}

// WriteAttributes writes a slice of Attributes x to Writer w.
func WriteAttributes(w *Writer, x *[]Attribute) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	for _, attribute := range *x {
		w.Float32(&attribute.Min)
		w.Float32(&attribute.Max)
		w.Float32(&attribute.Value)
		w.Float32(&attribute.Default)
		w.String(&attribute.Name)

		m := uint32(len(attribute.Modifiers))
		w.Varuint32(&m)
		for _, modifier := range attribute.Modifiers {
			w.String(&modifier.ID)
			w.String(&modifier.Name)
			w.Float32(&modifier.Amount)
			w.Int32(&modifier.Operation)
			w.Int32(&modifier.Operand)
			w.Bool(&modifier.Serializable)
		}
	}
}

// InitialAttributes reads an Attribute slice from bytes.Buffer src and stores it in the pointer passed.
// InitialAttributes is used when reading the attributes of a new entity. (AddEntity packet)
func InitialAttributes(r *Reader, x *[]Attribute) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)
	*x = make([]Attribute, count)
	for i := uint32(0); i < count; i++ {
		r.String(&(*x)[i].Name)
		r.Float32(&(*x)[i].Min)
		r.Float32(&(*x)[i].Value)
		r.Float32(&(*x)[i].Max)
	}
}

// WriteInitialAttributes writes a slice of Attributes x to Writer w. WriteInitialAttributes is used when
// writing the attributes of a new entity. (AddEntity packet)
func WriteInitialAttributes(w *Writer, x *[]Attribute) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	for _, attribute := range *x {
		w.String(&attribute.Name)
		w.Float32(&attribute.Min)
		w.Float32(&attribute.Value)
		w.Float32(&attribute.Max)
	}
}
