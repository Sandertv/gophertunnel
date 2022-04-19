package nbt

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
)

// Encoder writes NBT objects to an NBT output stream.
type Encoder struct {
	// Encoding is the variant to use for encoding the objects passed. By default, the variant is set to
	// NetworkLittleEndian, which is the variant used for network NBT.
	Encoding Encoding

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
	return &Encoder{w: writer, Encoding: NetworkLittleEndian}
}

// NewEncoderWithEncoding returns a new encoder for the output stream writer passed using a specific encoding.
// It is identical to calling NewEncoder and setting the Encoding field manually.
func NewEncoderWithEncoding(w io.Writer, encoding Encoding) *Encoder {
	enc := NewEncoder(w)
	enc.Encoding = encoding
	return enc
}

// Encode encodes an object to NBT and writes it to the NBT output stream of the encoder. See the Marshal
// docs for the conversion from Go types to NBT tags and special struct tags.
func (e *Encoder) Encode(v any) error {
	val := reflect.ValueOf(v)
	return e.marshal(val, "")
}

// Marshal encodes an object to its NBT representation and returns it as a byte slice. It uses the
// NetworkLittleEndian NBT encoding. To use a specific encoding, use MarshalEncoding.
//
// The Go value passed must be one of the values below, and structs, maps and slices may only have types that
// are found in the list below.
//
// The following Go types are converted to tags as such:
//   byte/uint8: TAG_Byte
//   bool: TAG_Byte
//   int16: TAG_Short
//   int32: TAG_Int
//   int64: TAG_Long
//   float32: TAG_Float
//   float64: TAG_Double
//   [...]byte: TAG_ByteArray
//   [...]int32: TAG_IntArray
//   [...]int64: TAG_LongArray
//   string: TAG_String
//   []<type>: TAG_List
//   struct{...}: TAG_Compound
//   map[string]<type/any>: TAG_Compound
//
// Marshal accepts struct fields with the 'nbt' struct tag. The 'nbt' struct tag allows setting the name of
// a field that some tag should be decoded in. Setting the struct tag to '-' means that field will never be
// filled by the decoding of the data passed. Suffixing the 'nbt' struct tag with ',omitempty' will prevent
// the field from being encoded if it is equal to its default value.
func Marshal(v any) ([]byte, error) {
	return MarshalEncoding(v, NetworkLittleEndian)
}

// MarshalEncoding encodes an object to its NBT representation and returns it as a byte slice. Its
// functionality is identical to that of Marshal, except it accepts any NBT encoding.
func MarshalEncoding(v any, encoding Encoding) ([]byte, error) {
	b := bufferPool.Get().(*bytes.Buffer)
	err := (&Encoder{w: &offsetWriter{Writer: b, WriteByte: b.WriteByte}, Encoding: encoding}).Encode(v)
	data := append([]byte(nil), b.Bytes()...)

	// Make sure to reset the buffer before putting it back in the pool.
	b.Reset()
	bufferPool.Put(b)
	return data, err
}

// bufferPool is a sync.Pool holding bytes.Buffers which are re-used for writing NBT, so that no new buffer
// needs to be allocated for each write.
var bufferPool = sync.Pool{
	New: func() any {
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
	if val.Kind() == reflect.Ptr {
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
	kind := val.Kind()
	if kind == reflect.Interface {
		val = val.Elem()
		kind = val.Kind()
	}
	switch vk := kind; vk {
	case reflect.Uint8:
		return e.w.WriteByte(byte(val.Uint()))

	case reflect.Bool:
		if val.Bool() {
			return e.w.WriteByte(1)
		}
		return e.w.WriteByte(0)

	case reflect.Int16:
		return e.Encoding.WriteInt16(e.w, int16(val.Int()))

	case reflect.Int32:
		return e.Encoding.WriteInt32(e.w, int32(val.Int()))

	case reflect.Int64:
		return e.Encoding.WriteInt64(e.w, val.Int())

	case reflect.Float32:
		return e.Encoding.WriteFloat32(e.w, float32(val.Float()))

	case reflect.Float64:
		return e.Encoding.WriteFloat64(e.w, val.Float())

	case reflect.Array:
		switch val.Type().Elem().Kind() {

		case reflect.Uint8:
			n := val.Cap()
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
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
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Encoding.WriteInt32(e.w, int32(val.Index(i).Int())); err != nil {
					return err
				}
			}

		case reflect.Int64:
			n := val.Cap()
			if err := e.Encoding.WriteInt32(e.w, int32(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := e.Encoding.WriteInt64(e.w, val.Index(i).Int()); err != nil {
					return err
				}
			}
		}

	case reflect.String:
		return e.Encoding.WriteString(e.w, val.String())

	case reflect.Slice:
		e.depth++
		elemType := val.Type().Elem()
		if elemType.Kind() == reflect.Interface {
			if val.Len() == 0 {
				// If the slice is empty, we cannot find out the type of the interface slice. Luckily the NBT
				// format allows a byte type for empty lists.
				elemType = byteType
			} else {
				// The slice is not empty, so we'll simply get the tag type from the first element.
				elemType = val.Index(0).Elem().Type()
			}
		}

		listType := tagFromType(elemType)
		if listType == math.MaxUint8 {
			return IncompatibleTypeError{Type: val.Type(), ValueName: tagName}
		}
		if err := e.w.WriteByte(listType); err != nil {
			return FailedWriteError{Off: e.w.off, Op: "WriteSlice", Err: err}
		}
		if err := e.Encoding.WriteInt32(e.w, int32(val.Len())); err != nil {
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
		fieldType := val.Type().Field(i)
		fieldValue := val.Field(i)
		tag := fieldType.Tag.Get("nbt")
		if fieldType.PkgPath != "" || tag == "-" {
			// Either the PkgPath was not empty, meaning we're dealing with an unexported field, or the
			// tag was '-', meaning we should skip it.
			continue
		}
		if fieldType.Anonymous {
			// The field was anonymous, so we write that in the same compound tag as this one.
			if err := e.writeStructValues(fieldValue); err != nil {
				return err
			}
			continue
		}
		tagName := fieldType.Name
		if strings.HasSuffix(tag, ",omitempty") {
			tag = strings.TrimSuffix(tag, ",omitempty")
			if reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()) {
				// The tag had the ',omitempty' tag, meaning it should be omitted if it has the zero
				// value. If this is reached, that was the case, and we skip it.
				continue
			}
		}
		if tag != "" {
			tagName = tag
		}
		if err := e.marshal(fieldValue, tagName); err != nil {
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
	return e.Encoding.WriteString(e.w, tagName)
}
