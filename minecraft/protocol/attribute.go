package protocol

import (
	"bytes"
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
}

// Attributes reads an Attribute slice from bytes.Buffer src and stores it in the pointer passed.
func Attributes(src *bytes.Buffer, attributes *[]Attribute) error {
	var count uint32
	if err := Varuint32(src, &count); err != nil {
		return wrap(err)
	}
	if count > mediumLimit {
		return LimitHitError{Limit: mediumLimit, Type: "attribute"}
	}
	*attributes = make([]Attribute, count)
	for i := uint32(0); i < count; i++ {
		attribute := Attribute{}
		if err := chainErr(
			Float32(src, &attribute.Min),
			Float32(src, &attribute.Max),
			Float32(src, &attribute.Value),
			Float32(src, &attribute.Default),
			String(src, &attribute.Name),
		); err != nil {
			return wrap(err)
		}
		(*attributes)[i] = attribute
	}
	return nil
}

// WriteAttributes writes a slice of Attributes x to buffer dst.
func WriteAttributes(dst *bytes.Buffer, x []Attribute) error {
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return wrap(err)
	}
	for _, attribute := range x {
		if err := chainErr(
			WriteFloat32(dst, attribute.Min),
			WriteFloat32(dst, attribute.Max),
			WriteFloat32(dst, attribute.Value),
			WriteFloat32(dst, attribute.Default),
			WriteString(dst, attribute.Name),
		); err != nil {
			return wrap(err)
		}
	}
	return nil
}

// InitialAttributes reads an Attribute slice from bytes.Buffer src and stores it in the pointer passed.
// InitialAttributes is used when reading the attributes of a new entity. (AddEntity packet)
func InitialAttributes(src *bytes.Buffer, attributes *[]Attribute) error {
	var count uint32
	if err := Varuint32(src, &count); err != nil {
		return wrap(err)
	}
	if count > mediumLimit {
		return LimitHitError{Limit: mediumLimit, Type: "attribute"}
	}
	*attributes = make([]Attribute, count)
	for i := uint32(0); i < count; i++ {
		attribute := Attribute{}
		if err := chainErr(
			String(src, &attribute.Name),
			Float32(src, &attribute.Min),
			Float32(src, &attribute.Value),
			Float32(src, &attribute.Max),
		); err != nil {
			return wrap(err)
		}
		(*attributes)[i] = attribute
	}
	return nil
}

// WriteInitialAttributes writes a slice of Attributes x to buffer dst. WriteInitialAttributes is used when
// writing the attributes of a new entity. (AddEntity packet)
func WriteInitialAttributes(dst *bytes.Buffer, x []Attribute) error {
	if err := WriteVaruint32(dst, uint32(len(x))); err != nil {
		return wrap(err)
	}
	for _, attribute := range x {
		if err := chainErr(
			WriteString(dst, attribute.Name),
			WriteFloat32(dst, attribute.Min),
			WriteFloat32(dst, attribute.Value),
			WriteFloat32(dst, attribute.Max),
		); err != nil {
			return wrap(err)
		}
	}
	return nil
}
