package nbt

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"reflect"
	"strings"
	"sync"
)

// Decoder reads NBT objects from an NBT input stream.
type Decoder struct {
	// Encoding is the variant to use for decoding the NBT passed. By default, the variant is set to
	// NetworkLittleEndian, which is the variant used for network NBT.
	Encoding Encoding
	// AllowZero, when set to true, prevents an error from being returned if the
	// first byte read from an io.Reader is 0x00 (TAG_End). This kind of data is
	// technically invalid, but some implementations do this to represent an
	// empty NBT tree.
	AllowZero bool

	r     *offsetReader
	depth int
}

// NewDecoder returns a new Decoder for the input stream reader passed.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{Encoding: NetworkLittleEndian, r: newOffsetReader(r)}
}

// NewDecoderWithEncoding returns a new Decoder for the input stream reader passed with a specific encoding.
func NewDecoderWithEncoding(r io.Reader, encoding Encoding) *Decoder {
	return &Decoder{Encoding: encoding, r: newOffsetReader(r)}
}

// Decode reads the next NBT object from the input stream and stores it into the pointer to an object passed.
// See the Unmarshal docs for the conversion between NBT tags to Go types.
func (d *Decoder) Decode(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return NonPointerTypeError{ActualType: val.Type()}
	}
	tagType, tagName, err := d.tag()
	if err != nil {
		return err
	}
	if tagType == tagEnd && d.AllowZero {
		return nil
	}
	return d.unmarshalTag(val.Elem(), tagType, tagName)
}

// Unmarshal decodes a slice of NBT data into a pointer to a Go values passed. Marshal will use the
// NetworkLittleEndian encoding by default. To use a specific encoding, use UnmarshalEncoding.
//
// The Go value passed must be a pointer to a value. Anything else will return an error before decoding.
// The following NBT tags are decoded in the Go value passed as such:
//
//	TAG_Byte: byte/uint8(/any) or bool
//	TAG_Short: int16(/any)
//	TAG_Int: int32(/any)
//	TAG_Long: int64(/any)
//	TAG_Float: float32(/any)
//	TAG_Double: float64(/any)
//	TAG_ByteArray: [...]byte(/any) (The value must be a byte array, not a slice)
//	TAG_String: string(/any)
//	TAG_List: []any(/any) (The value type of the slice may vary. Depending on the type of
//	          values in the List tag, it might be of the type of any of the other tags, such as []int64.
//
// TAG_Compound: struct{...}/map[string]any(/any)
// TAG_IntArray: [...]int32(/any) (The value must be an int32 array, not a slice)
// TAG_LongArray: [...]int64(/any) (The value must be an int64 array, not a slice)
//
// Unmarshal returns an error if the data is decoded into a struct and the struct does not have all fields
// that the matching TAG_Compound in the NBT has, in order to prevent the loss of data. For varying data, the
// data should be decoded into a map.
// Nil maps and slices are initialised and filled out automatically by Unmarshal.
//
// Unmarshal accepts struct fields with the 'nbt' struct tag. The 'nbt' struct tag allows setting the name of
// a field that some tag should be decoded in. Setting the struct tag to '-' means that field will never be
// filled by the decoding of the data passed.
func Unmarshal(data []byte, v any) error {
	return UnmarshalEncoding(data, v, NetworkLittleEndian)
}

// UnmarshalEncoding decodes a slice of NBT data into a pointer to a Go values passed using the NBT encoding
// passed. Its functionality is identical to that of Unmarshal, except that it allows a specific encoding.
func UnmarshalEncoding(data []byte, v any, encoding Encoding) error {
	buf := bytes.NewBuffer(data)
	return (&Decoder{Encoding: encoding, r: &offsetReader{
		Reader:   buf,
		ReadByte: buf.ReadByte,
		Next:     buf.Next,
	}}).Decode(v)
}

// These types are initialised once and re-used for each Unmarshal call.
var stringType = reflect.TypeOf("")
var byteType = reflect.TypeOf(byte(0))
var int32Type = reflect.TypeOf(int32(0))
var int64Type = reflect.TypeOf(int64(0))

// fieldMapPool is used to store maps holding the fields of a struct. These maps are cleared each time they
// are put back into the pool, but are re-used simply so that they need not to be re-allocated each operation.
var fieldMapPool = sync.Pool{
	New: func() any {
		return map[string]reflect.Value{}
	},
}

// unmarshalTag decodes a tag from the decoder's input stream into the reflect.Value passed, assuming the tag
// has the type and name passed.
func (d *Decoder) unmarshalTag(val reflect.Value, t tagType, tagName string) error {
	k := val.Kind()
	switch t {
	default:
		return UnknownTagError{Off: d.r.off, TagType: t, Op: "Match"}
	case tagEnd:
		return UnexpectedTagError{Off: d.r.off, TagType: tagEnd}
	case tagByte:
		value, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Byte"}
		}
		switch {
		case k == reflect.Uint8:
			val.SetUint(uint64(value))
		case k == reflect.Bool:
			val.SetBool(value == 1)
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagInt16:
		value, err := d.Encoding.Int16(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.Int16:
			val.SetInt(int64(value))
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagInt32:
		value, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.Int32:
			val.SetInt(int64(value))
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagInt64:
		value, err := d.Encoding.Int64(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.Int64:
			val.SetInt(value)
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagFloat32:
		value, err := d.Encoding.Float32(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.Float32:
			val.SetFloat(float64(value))
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagFloat64:
		value, err := d.Encoding.Float64(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.Float64:
			val.SetFloat(value)
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagString:
		value, err := d.Encoding.String(d.r)
		if err != nil {
			return err
		}
		switch {
		case k == reflect.String:
			val.SetString(value)
		case isAny(val):
			val.Set(reflect.ValueOf(value))
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
	case tagByteArray:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		b := make([]byte, length)
		if _, err := d.r.Read(b); err != nil {
			return BufferOverrunError{Op: "ByteArray"}
		}
		value := reflect.New(reflect.ArrayOf(int(length), byteType)).Elem()
		reflect.Copy(value, reflect.ValueOf(b))

		switch {
		case k == reflect.Array && val.Type().Elem().Kind() == reflect.Uint8:
			if val.Cap() != int(length) {
				return InvalidArraySizeError{Off: d.r.off, Op: "ByteArray", GoLength: val.Cap(), NBTLength: int(length)}
			}
		case isAny(val):
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
		val.Set(value)
	case tagInt32Array:
		s, err := d.Encoding.Int32Slice(d.r)
		if err != nil {
			return err
		}

		value := reflect.New(reflect.ArrayOf(len(s), int32Type)).Elem()
		reflect.Copy(value, reflect.ValueOf(s))

		switch {
		case k == reflect.Array && val.Type().Elem().Kind() == reflect.Int32:
			if val.Cap() != len(s) {
				return InvalidArraySizeError{Off: d.r.off, Op: "Int32Array", GoLength: val.Cap(), NBTLength: len(s)}
			}
		case isAny(val):
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
		val.Set(value)

	case tagInt64Array:
		s, err := d.Encoding.Int64Slice(d.r)
		if err != nil {
			return err
		}

		value := reflect.New(reflect.ArrayOf(len(s), int64Type)).Elem()
		reflect.Copy(value, reflect.ValueOf(s))

		switch {
		case k == reflect.Array && val.Type().Elem().Kind() == reflect.Int64:
			if val.Cap() != len(s) {
				return InvalidArraySizeError{Off: d.r.off, Op: "Int64Array", GoLength: val.Cap(), NBTLength: len(s)}
			}
		case isAny(val):
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
		val.Set(value)

	case tagSlice:
		d.depth++
		listTypeByte, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Slice"}
		}
		listType := tagType(listTypeByte)
		if !listType.IsValid() {
			return UnknownTagError{Off: d.r.off, TagType: listType, Op: "Slice"}
		}
		sliceType := val.Type()
		if val.Kind() != reflect.Slice && !isAny(val) {
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		}
		if val.Kind() == reflect.Interface {
			sliceType = reflect.SliceOf(sliceType)
		}
		switch listType {
		case tagByte:
			length, err := d.Encoding.Int32(d.r)
			if err != nil {
				return BufferOverrunError{Op: "ByteSlice"}
			}
			if length == 0 {
				// Empty lists are allowed to have the TAG_Byte type.
				val.Set(reflect.MakeSlice(sliceType, int(length), int(length)))
				break
			}
			b := make([]byte, length)
			if _, err := d.r.Read(b); err != nil {
				return BufferOverrunError{Op: "ByteSlice"}
			}
			switch {
			case k == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8, isAny(val):
				val.Set(reflect.ValueOf(b))
			default:
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type().Elem(), Field: tagName, TagType: listType}
			}
		case tagInt32:
			b, err := d.Encoding.Int32Slice(d.r)
			if err != nil {
				return BufferOverrunError{Op: "Int32Slice"}
			}
			switch {
			case k == reflect.Slice && val.Type().Elem().Kind() == reflect.Int32, isAny(val):
				val.Set(reflect.ValueOf(b))
			default:
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type().Elem(), Field: tagName, TagType: listType}
			}
		case tagInt64:
			b, err := d.Encoding.Int64Slice(d.r)
			if err != nil {
				return BufferOverrunError{Op: "Int64Slice"}
			}
			switch {
			case k == reflect.Slice && val.Type().Elem().Kind() == reflect.Int64, isAny(val):
				val.Set(reflect.ValueOf(b))
			default:
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type().Elem(), Field: tagName, TagType: listType}
			}
		default:
			length, err := d.Encoding.Int32(d.r)
			if err != nil {
				return err
			}
			v := reflect.MakeSlice(sliceType, int(length), int(length))
			for i := 0; i < int(length); i++ {
				if err := d.unmarshalTag(v.Index(i), listType, ""); err != nil {
					// An error occurred during the decoding of one of the elements of the TAG_List, meaning it
					// either had an invalid type or the NBT was invalid.
					if e, ok := err.(InvalidTypeError); ok {
						return InvalidTypeError{Off: d.r.off, FieldType: sliceType.Elem(), Field: fmt.Sprintf("%v[%v].%v", tagName, i, e.Field), TagType: listType}
					}
					return err
				}
			}
			val.Set(v)
			d.depth--
		}

	case tagStruct:
		d.depth++
		switch val.Kind() {
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
		case reflect.Struct:
			// We first fetch a fields map from the sync.Pool. These maps already have a base size obtained
			// from when they were used, meaning we don't have to re-allocate each element.
			fields := fieldMapPool.Get().(map[string]reflect.Value)
			d.populateFields(val, fields)
			for {
				nestedTagType, nestedTagName, err := d.tag()
				if err != nil {
					return err
				}
				if nestedTagType == tagEnd {
					// We reached the end of the fields.
					break
				}
				if !nestedTagType.IsValid() {
					return UnknownTagError{Off: d.r.off, Op: "Struct", TagType: nestedTagType}
				}
				field, ok := fields[nestedTagName]
				if ok {
					if err = d.unmarshalTag(field, nestedTagType, nestedTagName); err != nil {
						return err
					}
					continue
				}
				// We return an error if the struct does not have one of the fields found in the compound. It
				// is rather important no data is lost during the decoding.
				return UnexpectedNamedTagError{Off: d.r.off, TagName: tagName + "." + nestedTagName, TagType: nestedTagType}
			}
			// Finally we delete all fields in the map and return it to the sync.Pool so that it may be
			// re-used by the next operation.
			for k := range fields {
				delete(fields, k)
			}
			fieldMapPool.Put(fields)
		case reflect.Interface, reflect.Map:
			if vk := val.Kind(); vk == reflect.Interface && val.NumMethod() != 0 {
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: t}
			}
			valType := val.Type()
			if val.Kind() == reflect.Map {
				valType = valType.Elem()
			}
			m := reflect.MakeMap(reflect.MapOf(stringType, valType))
			for {
				nestedTagType, nestedTagName, err := d.tag()
				if err != nil {
					return err
				}
				if !nestedTagType.IsValid() {
					return UnknownTagError{Off: d.r.off, Op: "Map", TagType: nestedTagType}
				}
				if nestedTagType == tagEnd {
					// We reached the end of the compound.
					break
				}
				value := reflect.New(valType).Elem()
				if err := d.unmarshalTag(value, nestedTagType, nestedTagName); err != nil {
					return err
				}
				m.SetMapIndex(reflect.ValueOf(nestedTagName), value)
			}
			val.Set(m)
		}
		d.depth--
	}
	return nil
}

// populateFields populates the map passed with the fields of the reflect representation of a struct passed.
// It takes into consideration the nbt struct field tag.
func (d *Decoder) populateFields(val reflect.Value, m map[string]reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		if !ast.IsExported(fieldType.Name) {
			// The struct field's name was not exported.
			continue
		}
		field := val.Field(i)
		name := fieldType.Name
		if fieldType.Anonymous {
			// We got an anonymous struct field, so we decode that into the same level.
			d.populateFields(field, m)
			continue
		}
		if tag, ok := fieldType.Tag.Lookup("nbt"); ok {
			if tag == "-" {
				continue
			}
			tag = strings.TrimSuffix(tag, ",omitempty")
			if tag != "" {
				name = tag
			}
		}
		m[name] = field
	}
}

// tag reads a tag from the decoder, and its name if the tag type is not a TAG_End.
func (d *Decoder) tag() (t tagType, tagName string, err error) {
	if d.depth >= maximumNestingDepth {
		return 0, "", MaximumDepthReachedError{}
	}
	if d.r.off >= maximumNetworkOffset && d.Encoding == NetworkLittleEndian {
		return 0, "", MaximumBytesReadError{}
	}
	tagTypeByte, err := d.r.ReadByte()
	if err != nil {
		return 0, "", BufferOverrunError{Op: "ReadTag"}
	}
	t = tagType(tagTypeByte)
	if _, ok := d.Encoding.(networkBigEndian); ok && t == tagStruct && d.depth == 0 {
		// As of Minecraft Java 1.20.2, the name of the root compound tag is not written over the network.
		return t, "", err
	}
	if t != tagEnd {
		// Only read a tag name if the tag's type is not TAG_End.
		tagName, err = d.Encoding.String(d.r)
	}
	return t, tagName, err
}

// isAny checks if a reflect.Value has the type `any` or `interface{}`.
func isAny(v reflect.Value) bool {
	return v.Kind() == reflect.Interface && v.NumMethod() == 0
}
