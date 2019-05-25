package nbt

import (
	"bytes"
	"errors"
	"io"
	"math"
	"reflect"
	"sync"
	"unicode/utf8"
)

// Encoder writes NBT objects to an NBT output stream.
type Encoder struct {
	// Variant is the variant to use for encoding the objects passed. By default, the variant is set to
	// VarLittleEndian, which is the variant used for network NBT.
	Variant Variant

	w     *offsetWriter
	depth int
}

// NewEncoder returns a new encoder for the output stream writer passed.
func NewEncoder(w io.Writer) *Encoder {
	var writer *offsetWriter
	if byteWriter, ok := w.(io.ByteWriter); ok {
		writer = &offsetWriter{Writer: w, WriteByte: byteWriter.WriteByte}
	} else {
		writer = &offsetWriter{Writer: w, WriteByte: func(b byte) error {
			_, err := w.Write([]byte{b})
			return err
		}}
	}
	return &Encoder{w: writer, Variant: VarLittleEndian}
}

// Encode encodes an object to NBT and writes it to the NBT output stream of the encoder. See the Marshal
// docs for the conversion from Go types to NBT tags and special struct tags.
func (e *Encoder) Encode(v interface{}) error {
	val := reflect.ValueOf(v)
	return e.marshal(val, "")
}

// Marshal encodes an object to its NBT representation and returns it as a byte slice. It uses the
// VarLittleEndian NBT format. To use a specific format, use MarshalVariant.
//
// The Go value passed must be one of the values below, and structs, maps and slices may only have types that
// are found in the list below.
//
// The following Go types are converted to tags as such:
// byte/uint8: TAG_Byte
// int16: TAG_Short
// int32: TAG_Int
// int64: TAG_Long
// float32: TAG_Float
// float64: TAG_Double
// [...]byte: TAG_ByteArray
// [...]int32: TAG_IntArray
// [...]int64: TAG_LongArray
// string: TAG_String
// []<type>: TAG_List
// struct{...}: TAG_Compound
// map[string]<type/interface{}>: TAG_Compound
//
// Marshal accepts struct fields with the 'nbt' struct tag. The 'nbt' struct tag allows setting the name of
// a field that some tag should be decoded in. Setting the struct tag to '-' means that field will never be
// filled by the decoding of the data passed. Suffixing the 'nbt' struct tag with ',omitempty' will prevent
// the field from being encoded if it is equal to its default value.
func Marshal(v interface{}) ([]byte, error) {
	return MarshalVariant(v, VarLittleEndian)
}

// MarshalVariant encodes an object to its NBT representation and returns it as a byte slice. Its
// functionality is identical to that of Marshal, except it accepts any NBT variant.
func MarshalVariant(v interface{}, variant Variant) ([]byte, error) {
	b := bufferPool.Get().(*bytes.Buffer)
	err := (&Encoder{w: &offsetWriter{Writer: b, WriteByte: b.WriteByte}, Variant: variant}).Encode(v)
	data := append([]byte(nil), b.Bytes()...)

	// Make sure to reset the buffer before putting it back in the pool.
	b.Reset()
	bufferPool.Put(b)
	return data, err
}

// bufferPool is a sync.Pool holding bytes.Buffers which are re-used for writing NBT, so that no new buffer
// needs to be allocated for each write.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 64))
	},
}

// marshal encodes a reflect.Value with a tag name passed to its NBT representation. It writes the tag type,
// name and payload. An error is returned if any values in the reflect.Value found were not representable
// with an NBT tag.
func (e *Encoder) marshal(val reflect.Value, tagName string) error {
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	tagType := tagFromType(val.Type())
	if tagType == math.MaxUint8 {
		return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
	}
	if err := e.writeTag(tagType, tagName); err != nil {
		return err
	}
	return e.encode(val, tagName)
}

// encode encodes the payload of a value passed with the tag name passed. Unlike calling Encoder.marshal(), it
// does not write the name and type of the tag.
func (e *Encoder) encode(val reflect.Value, tagName string) error {
	switch vk := val.Kind(); vk {
	case reflect.Uint8:
		return e.w.WriteByte(byte(val.Uint()))

	case reflect.Int16:
		return e.Variant.WriteInt16(e.w, int16(val.Int()))

	case reflect.Int32:
		return e.Variant.WriteInt32(e.w, int32(val.Int()))

	case reflect.Int64:
		return e.Variant.WriteInt64(e.w, int64(val.Int()))

	case reflect.Float32:
		return e.Variant.WriteFloat32(e.w, float32(val.Float()))

	case reflect.Float64:
		return e.Variant.WriteFloat64(e.w, val.Float())

	case reflect.Array:
		switch val.Type().Elem().Kind() {

		case reflect.Uint8:
			n := val.Cap()
			if err := e.Variant.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			data := make([]byte, n)
			for i := 0; i < n; i++ {
				data[i] = byte(val.Index(i).Uint())
			}
			if _, err := e.w.Write(data); err != nil {
				return FailedWriteError{Op: "WriteByteArray", Off: e.w.off}
			}
			return nil

		case reflect.Int32:
			n := val.Cap()
			if err := e.Variant.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Variant.WriteInt32(e.w, int32(val.Index(i).Int())); err != nil {
					return err
				}
			}

		case reflect.Int64:
			n := val.Cap()
			if err := e.Variant.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Variant.WriteInt64(e.w, val.Index(i).Int()); err != nil {
					return err
				}
			}
		}

	case reflect.String:
		s := val.String()
		if !utf8.ValidString(s) {
			return InvalidStringError{Off: e.r.off, String: s, Err: errors.New("string does not exist out of utf8 only")}
		}
		return e.Variant.WriteString(e.w, s)

	case reflect.Slice:
		e.depth++
		listType := tagFromType(val.Type().Elem())
		if listType == math.MaxUint8 {
			return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
		}
		if err := e.w.WriteByte(listType); err != nil {
			return FailedWriteError{Off: e.w.off, Op: "WriteSlice", Err: err}
		}
		if err := e.Variant.WriteInt32(e.w, int32(val.Len())); err != nil {
			return err
		}
		for i := 0; i < val.Len(); i++ {
			nestedValue := val.Index(i)
			if err := e.encode(nestedValue, ""); err != nil {
				return err
			}
		}
		e.depth--

	case reflect.Struct:
		e.depth++
		if err := e.writeStructValues(val); err != nil {
			return err
		}
		e.depth--
		return e.w.WriteByte(tagEnd)

	case reflect.Map:
		e.depth++
		if val.Type().Key().Kind() != reflect.String {
			return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
		}
		iter := val.MapRange()
		for iter.Next() {
			if err := e.marshal(iter.Value(), iter.Key().String()); err != nil {
				return err
			}
		}
		e.depth--
		return e.w.WriteByte(tagEnd)
	}
	return nil
}

// writeStructValues writes the values of all struct fields of a reflect.Value (must be of struct type) to
// the io.Writer of the encoder.
func (e *Encoder) writeStructValues(val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		valType := val.Type().Field(i)
		valValue := val.Field(i)
		tag := valType.Tag.Get("nbt")
		if valType.PkgPath != "" || tag == "-" {
			// Either the PkgPath was not empty, meaning we're dealing with an unexported field, or the
			// tag was '-', meaning we should skip it.
			continue
		}
		if valType.Anonymous {
			// The field was anonymous, so we write that in the same compound tag as this one.
			if err := e.writeStructValues(valValue); err != nil {
				return err
			}
			continue
		}
		tagName := valType.Name
		if tag != "" {
			omitEmptyLen := len(",omitempty")
			tagLen := len(tag)
			// Make sure the tag's length is at least as long as ',omitempty'
			if tagLen >= omitEmptyLen {
				omitEmpty := tag[tagLen-omitEmptyLen:]
				if omitEmpty == ",omitempty" {
					if reflect.DeepEqual(valValue, reflect.Zero(valValue.Type())) {
						// The tag had the ',omitempty' tag, meaning it should be omitted if it has the zero
						// value. If this is reached, that was the case, and we skip it.
						continue
					}
					tag = tag[:tagLen-omitEmptyLen]
				}
			}
			tagName = tag
		}
		if err := e.marshal(valValue, tagName); err != nil {
			return err
		}
	}
	return nil
}

// writeTag writes a single tag to the io.Writer held by the Encoder. The tag type and the name are written.
func (e *Encoder) writeTag(tagType byte, tagName string) error {
	if e.depth >= maximumNestingDepth {
		return MaximumDepthReachedError{}
	}
	if err := e.w.WriteByte(tagType); err != nil {
		return err
	}
	return e.Variant.WriteString(e.w, tagName)
}
