package nbt

import (
	"math"
	"reflect"
)

const (
	tagEnd byte = iota
	tagByte
	tagInt16
	tagInt32
	tagInt64
	tagFloat32
	tagFloat64
	tagByteArray
	tagString
	tagSlice
	tagStruct
	tagInt32Array
	tagInt64Array
)

// tagName returns the name of a tag. The function panics if the tag does not exist.
func tagName(tag byte) string {
	switch tag {
	case tagEnd:
		return "TAG_End"
	case tagByte:
		return "TAG_Byte"
	case tagInt16:
		return "TAG_Short"
	case tagInt32:
		return "TAG_Int"
	case tagInt64:
		return "TAG_Long"
	case tagFloat32:
		return "TAG_Float"
	case tagFloat64:
		return "TAG_Double"
	case tagByteArray:
		return "TAG_ByteArray"
	case tagString:
		return "TAG_String"
	case tagSlice:
		return "TAG_List"
	case tagStruct:
		return "TAG_Compound"
	case tagInt32Array:
		return "TAG_IntArray"
	case tagInt64Array:
		return "TAG_LongArray"
	default:
		panic("unknown tag")
	}
}

// tagExists checks if a tag exists.
func tagExists(tag byte) bool {
	switch tag {
	case tagEnd, tagByte, tagInt16, tagInt32, tagInt64, tagFloat32, tagFloat64, tagByteArray, tagString,
		tagSlice, tagStruct, tagInt32Array, tagInt64Array:
		return true
	default:
		return false
	}
}

// tagFromType matches a reflect.Type with a tag type that can hold its value. If none is found, math.MaxUint8
// is returned.
func tagFromType(p reflect.Type) byte {
	switch p.Kind() {
	case reflect.Uint8, reflect.Bool:
		return tagByte
	case reflect.Int16:
		return tagInt16
	case reflect.Int32:
		return tagInt32
	case reflect.Int64:
		return tagInt64
	case reflect.Float32:
		return tagFloat32
	case reflect.Float64:
		return tagFloat64
	case reflect.Array:
		switch p.Elem().Kind() {
		case reflect.Uint8:
			return tagByteArray
		case reflect.Int32:
			return tagInt32Array
		case reflect.Int64:
			return tagInt64Array
		}
	case reflect.String:
		return tagString
	case reflect.Slice:
		return tagSlice
	case reflect.Struct, reflect.Map:
		return tagStruct
	}
	return math.MaxUint8
}
