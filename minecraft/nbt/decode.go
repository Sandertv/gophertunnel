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
	return d.unmarshalTag(val.Elem(), tagType, tagName)
}

// Unmarshal decodes a slice of NBT data into a pointer to a Go values passed. Marshal will use the
// NetworkLittleEndian encoding by default. To use a specific encoding, use UnmarshalEncoding.
//
// The Go value passed must be a pointer to a value. Anything else will return an error before decoding.
// The following NBT tags are decoded in the Go value passed as such:
//   TAG_Byte: byte/uint8(/any) or bool
//   TAG_Short: int16(/any)
//   TAG_Int: int32(/any)
//   TAG_Long: int64(/any)
//   TAG_Float: float32(/any)
//   TAG_Double: float64(/any)
//   TAG_ByteArray: [...]byte(/any) (The value must be a byte array, not a slice)
//   TAG_String: string(/any)
//   TAG_List: []any(/any) (The value type of the slice may vary. Depending on the type of
//             values in the List tag, it might be of the type of any of the other tags, such as []int64.
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
func (d *Decoder) unmarshalTag(val reflect.Value, tagType byte, tagName string) error {
	switch tagType {
	default:
		return UnknownTagError{Off: d.r.off, TagType: tagType, Op: "Match"}
	case tagEnd:
		return UnexpectedTagError{Off: d.r.off, TagType: tagEnd}

	case tagByte:
		value, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "Byte"}
		}
		if val.Kind() != reflect.Uint8 {
			if val.Kind() == reflect.Bool {
				if value != 0 {
					val.SetBool(true)
				}
				return nil
			}
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetUint(uint64(value))

	case tagInt16:
		value, err := d.Encoding.Int16(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int16 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(int64(value))

	case tagInt32:
		value, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(int64(value))

	case tagInt64:
		value, err := d.Encoding.Int64(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Int64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetInt(value)

	case tagFloat32:
		value, err := d.Encoding.Float32(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Float32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetFloat(float64(value))

	case tagFloat64:
		value, err := d.Encoding.Float64(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.Float64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetFloat(value)

	case tagString:
		value, err := d.Encoding.String(d.r)
		if err != nil {
			return err
		}
		if val.Kind() != reflect.String {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(reflect.ValueOf(value))
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		val.SetString(value)

	case tagByteArray:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		data, err := consumeN(int(length), d.r)
		if err != nil {
			return BufferOverrunError{Op: "ByteArray"}
		}
		value := reflect.New(reflect.ArrayOf(int(length), byteType)).Elem()
		for i := int32(0); i < length; i++ {
			value.Index(int(i)).SetUint(uint64(data[i]))
		}
		if val.Kind() != reflect.Array {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "ByteArray", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagInt32Array:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		value := reflect.New(reflect.ArrayOf(int(length), int32Type)).Elem()
		for i := int32(0); i < length; i++ {
			v, err := d.Encoding.Int32(d.r)
			if err != nil {
				return err
			}
			value.Index(int(i)).SetInt(int64(v))
		}
		if val.Kind() != reflect.Array || val.Type().Elem().Kind() != reflect.Int32 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "Int32Array", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagInt64Array:
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		value := reflect.New(reflect.ArrayOf(int(length), int64Type)).Elem()
		for i := int32(0); i < length; i++ {
			v, err := d.Encoding.Int64(d.r)
			if err != nil {
				return err
			}
			value.Index(int(i)).SetInt(v)
		}
		if val.Kind() != reflect.Array || val.Type().Elem().Kind() != reflect.Int64 {
			if val.Kind() == reflect.Interface && val.NumMethod() == 0 {
				// Empty interface.
				val.Set(value)
				return nil
			}
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Cap() != int(length) {
			return InvalidArraySizeError{Off: d.r.off, Op: "Int64Array", GoLength: val.Cap(), NBTLength: int(length)}
		}
		val.Set(value)

	case tagSlice:
		d.depth++
		listType, err := d.r.ReadByte()
		if err != nil {
			return BufferOverrunError{Op: "List"}
		}
		if !tagExists(listType) {
			return UnknownTagError{Off: d.r.off, TagType: listType, Op: "Slice"}
		}
		length, err := d.Encoding.Int32(d.r)
		if err != nil {
			return err
		}
		valType := val.Type()
		if val.Kind() != reflect.Slice && val.Kind() != reflect.Interface {
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
		}
		if val.Kind() == reflect.Interface {
			valType = reflect.SliceOf(valType)
		}
		v := reflect.MakeSlice(valType, int(length), int(length))
		if length != 0 {
			for i := 0; i < int(length); i++ {
				if err := d.unmarshalTag(v.Index(i), listType, ""); err != nil {
					// An error occurred during the decoding of one of the elements of the TAG_List, meaning it
					// either had an invalid type or the NBT was invalid.
					if _, ok := err.(InvalidTypeError); ok {
						return InvalidTypeError{Off: d.r.off, FieldType: valType.Elem(), Field: fmt.Sprintf("%v[%v]", tagName, i), TagType: listType}
					}
					return err
				}
			}
		}
		val.Set(v)
		d.depth--

	case tagStruct:
		d.depth++
		switch val.Kind() {
		default:
			return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
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
				if !tagExists(nestedTagType) {
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
				return InvalidTypeError{Off: d.r.off, FieldType: val.Type(), Field: tagName, TagType: tagType}
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
				if !tagExists(nestedTagType) {
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
func (d *Decoder) tag() (tagType byte, tagName string, err error) {
	if d.depth >= maximumNestingDepth {
		return 0, "", MaximumDepthReachedError{}
	}
	if d.r.off >= maximumNetworkOffset && d.Encoding == NetworkLittleEndian {
		return 0, "", MaximumBytesReadError{}
	}
	tagType, err = d.r.ReadByte()
	if err != nil {
		return 0, "", BufferOverrunError{Op: "ReadTag"}
	}
	if tagType != tagEnd {
		// Only read a tag name if the tag's type is not TAG_End.
		tagName, err = d.Encoding.String(d.r)
	}
	return
}
