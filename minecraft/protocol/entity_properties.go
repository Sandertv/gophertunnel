package protocol

// EntityProperties holds lists of entity properties that define specific attributes of an entity. As of v1.19.40, the
// vanilla server does not use these properties, however they are still supported by the protocol.
type EntityProperties struct {
	// IntegerProperties is a list of entity properties that contain integer values.
	IntegerProperties []IntegerEntityProperty
	// FloatProperties is a list of entity properties that contain float values.
	FloatProperties []FloatEntityProperty
}

// Marshal ...
func (e *EntityProperties) Marshal(r IO) {
	Slice(r, &e.IntegerProperties)
	Slice(r, &e.FloatProperties)
}

// IntegerEntityProperty is an entity property that contains an integer value.
type IntegerEntityProperty struct {
	// Index represents the index of the property. It is unclear what the exact purpose of this is.
	Index uint32
	// Value is the value of the property.
	Value int32
}

// Marshal ...
func (i *IntegerEntityProperty) Marshal(r IO) {
	r.Varuint32(&i.Index)
	r.Varint32(&i.Value)
}

// FloatEntityProperty is an entity property that contains a float value.
type FloatEntityProperty struct {
	// Index represents the index of the property. It is unclear what the exact purpose of this is.
	Index uint32
	// Value is the value of the property.
	Value float32
}

// Marshal ...
func (f *FloatEntityProperty) Marshal(r IO) {
	r.Varuint32(&f.Index)
	r.Float32(&f.Value)
}
