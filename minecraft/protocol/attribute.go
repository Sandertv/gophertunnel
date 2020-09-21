package protocol

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
}

// Attributes reads an Attribute slice x from Reader r.
func Attributes(r *Reader, x *[]Attribute) {
	var count uint32
	r.Varuint32(&count)
	r.LimitUint32(count, mediumLimit)

	*x = make([]Attribute, count)
	for i := uint32(0); i < count; i++ {
		r.Float32(&(*x)[i].Min)
		r.Float32(&(*x)[i].Max)
		r.Float32(&(*x)[i].Value)
		r.Float32(&(*x)[i].Default)
		r.String(&(*x)[i].Name)
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
