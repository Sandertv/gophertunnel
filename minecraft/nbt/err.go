package nbt

import (
	"fmt"
	"reflect"
)

// InvalidTypeError is returned when the type of a tag read is not equal to the struct field with the name
// of that tag.
type InvalidTypeError struct {
	Off       int64
	Field     string
	TagType   byte
	FieldType reflect.Type
}

// Error ...
func (err InvalidTypeError) Error() string {
	return fmt.Sprintf("nbt: invalid type for tag '%v' at offset %v: cannot unmarshalTag %v into %v", err.Field, err.Off, tagName(err.TagType), err.FieldType)
}

// UnknownTagError is returned when the type of a tag read is not known, meaning it is not found in the tag.go
// file.
type UnknownTagError struct {
	Off     int64
	Op      string
	TagType byte
}

// Error ...
func (err UnknownTagError) Error() string {
	return fmt.Sprintf("nbt: unknown tag '%v' at offset %v during op '%v'", err.TagType, err.Off, err.Op)
}

// UnexpectedTagError is returned when a tag type encountered was not expected, and thus valid in its context.
type UnexpectedTagError struct {
	Off     int64
	TagType byte
}

// Error ...
func (err UnexpectedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected tag %v at offset %v: tag is not valid in its context", tagName(err.TagType), err.Off)
}

// NonPointerTypeError is returned when the type of a value passed in Decoder.Decode or Unmarshal is not a
// pointer.
type NonPointerTypeError struct {
	ActualType reflect.Type
}

// Error ...
func (err NonPointerTypeError) Error() string {
	return fmt.Sprintf("nbt: expected ptr type to decode into, but got '%v'", err.ActualType)
}

// BufferOverrunError is returned when the data buffer passed in when reading is overrun, meaning one of the
// reading operations extended beyond the end of the slice.
type BufferOverrunError struct {
	Op string
}

// Error ...
func (err BufferOverrunError) Error() string {
	return fmt.Sprintf("nbt: unexpected buffer end during op: '%v'", err.Op)
}

// InvalidArraySizeError is returned when an array read from the NBT (that includes byte arrays, int32 arrays
// and int64 arrays) does not have the same size as the Go representation.
type InvalidArraySizeError struct {
	Off       int64
	Op        string
	GoLength  int
	NBTLength int
}

// Error ...
func (err InvalidArraySizeError) Error() string {
	return fmt.Sprintf("nbt: mismatched array size at %v during op '%v': expected size %v, found %v in NBT", err.Off, err.Op, err.GoLength, err.NBTLength)
}

// UnexpectedNamedTagError is returned when a named tag is read from a compound which is not present in the
// struct it is decoded into.
type UnexpectedNamedTagError struct {
	Off     int64
	TagName string
	TagType byte
}

// Error ...
func (err UnexpectedNamedTagError) Error() string {
	return fmt.Sprintf("nbt: unexpected named tag '%v' with type %v at offset %v: not present in struct to be decoded into", err.TagName, tagName(err.TagType), err.Off)
}

// FailedWriteError is returned if a Write operation failed on an offsetWriter, meaning some of the data could
// not be written to the io.Writer.
type FailedWriteError struct {
	Off int64
	Op  string
	Err error
}

// Error ...
func (err FailedWriteError) Error() string {
	return fmt.Sprintf("nbt: failed write during op '%v' at offset %v: %v", err.Op, err.Off, err.Err)
}

// IncompatibleTypeError is returned if a value is attempted to be written to an io.Writer, but its type can-
// not be translated to an NBT tag.
type IncompatibleTypeError struct {
	ValueName string
	Type      reflect.Type
}

// Error ...
func (err IncompatibleTypeError) Error() string {
	return fmt.Sprintf("nbt: value type %v (%v) cannot be translated to an NBT tag", err.Type, err.ValueName)
}

// InvalidStringError is returned if a string read is not valid, meaning it does not exist exclusively out of
// utf8 characters, or if it is longer than the length prefix can carry.
type InvalidStringError struct {
	Off    int64
	Err    error
	String string
}

// Error ...
func (err InvalidStringError) Error() string {
	return fmt.Sprintf("nbt: string at offset %v is not valid: %v (%v)", err.Off, err.Err, err.String)
}

const maximumNestingDepth = 512

// MaximumDepthReachedError is returned if the maximum depth of 512 compound/list tags has been reached while
// reading or writing NBT.
type MaximumDepthReachedError struct {
}

// Error ...
func (err MaximumDepthReachedError) Error() string {
	return fmt.Sprintf("nbt: maximum nesting depth of %v was reached", maximumNestingDepth)
}

const maximumNetworkOffset = 4 * 1024 * 1024

// MaximumBytesReadError is returned if the maximum amount of bytes has been read for NetworkLittleEndian
// format. It is returned if the offset hits maximumNetworkOffset.
type MaximumBytesReadError struct {
}

// Error ...
func (err MaximumBytesReadError) Error() string {
	return fmt.Sprintf("nbt: limit of bytes read %v with NetworkLittleEndian format exhausted", maximumNetworkOffset)
}
