package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"image/color"
	"io"
	"math"
	"math/big"
	"math/bits"
	"unsafe"
)

// Reader implements reading operations for reading types from Minecraft packets. Each Packet implementation
// has one passed to it.
// Reader's uses should always be encapsulated with a deferred recovery. Reader panics on invalid data.
type Reader struct {
	r interface {
		io.Reader
		io.ByteReader
	}
	shieldID      int32
	limitsEnabled bool
}

// NewReader creates a new Reader using the io.ByteReader passed as underlying source to read bytes from.
func NewReader(r interface {
	io.Reader
	io.ByteReader
}, shieldID int32, enableLimits bool) *Reader {
	return &Reader{r: r, shieldID: shieldID, limitsEnabled: enableLimits}
}

// Uint8 reads a uint8 from the underlying buffer.
func (r *Reader) Uint8(x *uint8) {
	var err error
	*x, err = r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
}

// Int8 reads an int8 from the underlying buffer.
func (r *Reader) Int8(x *int8) {
	var b uint8
	r.Uint8(&b)
	*x = int8(b)
}

// Bool reads a bool from the underlying buffer.
func (r *Reader) Bool(x *bool) {
	u, err := r.r.ReadByte()
	if err != nil {
		r.panic(err)
	}
	*x = *(*bool)(unsafe.Pointer(&u))
}

// errStringTooLong is an error set if a string decoded using the String method has a length that is too long.
var errStringTooLong = errors.New("string length overflows a 32-bit integer")

// StringUTF ...
func (r *Reader) StringUTF(x *string) {
	var length int16
	r.Int16(&length)
	l := int(length)
	if l > math.MaxInt16 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

// String reads a string from the underlying buffer.
func (r *Reader) String(x *string) {
	var length uint32
	r.Varuint32(&length)
	l := int(length)
	if l > math.MaxInt32 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = *(*string)(unsafe.Pointer(&data))
}

// ByteSlice reads a byte slice from the underlying buffer, similarly to String.
func (r *Reader) ByteSlice(x *[]byte) {
	var length uint32
	r.Varuint32(&length)
	l := int(length)
	if l > math.MaxInt32 {
		r.panic(errStringTooLong)
	}
	data := make([]byte, l)
	if _, err := r.r.Read(data); err != nil {
		r.panic(err)
	}
	*x = data
}

// Vec3 reads three float32s into an mgl32.Vec3 from the underlying buffer.
func (r *Reader) Vec3(x *mgl32.Vec3) {
	r.Float32(&x[0])
	r.Float32(&x[1])
	r.Float32(&x[2])
}

// Vec2 reads two float32s into an mgl32.Vec2 from the underlying buffer.
func (r *Reader) Vec2(x *mgl32.Vec2) {
	r.Float32(&x[0])
	r.Float32(&x[1])
}

// BlockPos reads three varint32s into a BlockPos from the underlying buffer.
func (r *Reader) BlockPos(x *BlockPos) {
	r.Varint32(&x[0])
	r.Varint32(&x[1])
	r.Varint32(&x[2])
}

// UBlockPos reads three varint32s, one unsigned for the y, into a BlockPos from the underlying buffer.
func (r *Reader) UBlockPos(x *BlockPos) {
	r.Varint32(&x[0])
	var y uint32
	r.Varuint32(&y)
	x[1] = int32(y)
	r.Varint32(&x[2])
}

// ChunkPos writes a ChunkPos as 2 varint32s to the underlying buffer.
func (r *Reader) ChunkPos(x *ChunkPos) {
	r.Varint32(&x[0])
	r.Varint32(&x[1])
}

// SubChunkPos writes a SubChunkPos as 3 varint32s to the underlying buffer.
func (r *Reader) SubChunkPos(x *SubChunkPos) {
	r.Varint32(&x[0])
	r.Varint32(&x[1])
	r.Varint32(&x[2])
}

// SoundPos reads an mgl32.Vec3 that serves as a position for a sound.
func (r *Reader) SoundPos(x *mgl32.Vec3) {
	var b BlockPos
	r.UBlockPos(&b)
	*x = mgl32.Vec3{float32(b[0]) / 8, float32(b[1]) / 8, float32(b[2]) / 8}
}

// ByteFloat reads a rotational float32 from a single byte.
func (r *Reader) ByteFloat(x *float32) {
	var v uint8
	r.Uint8(&v)
	*x = float32(v) * (360.0 / 256.0)
}

// RGB reads a color.RGBA x from three float32s.
func (r *Reader) RGB(x *color.RGBA) {
	var red, green, blue float32
	r.Float32(&red)
	r.Float32(&green)
	r.Float32(&blue)
	*x = color.RGBA{
		R: uint8(red * 255),
		G: uint8(green * 255),
		B: uint8(blue * 255),
	}
}

// RGBA reads a color.RGBA x from a uint32.
func (r *Reader) RGBA(x *color.RGBA) {
	var v uint32
	r.Uint32(&v)
	*x = color.RGBA{
		R: byte(v),
		G: byte(v >> 8),
		B: byte(v >> 16),
		A: byte(v >> 24),
	}
}

// ARGB reads a color.ARGB x from a int32.
func (r *Reader) ARGB(x *color.RGBA) {
	var v int32
	r.Int32(&v)
	*x = color.RGBA{
		A: byte(v),
		R: byte(v >> 8),
		G: byte(v >> 16),
		B: byte(v >> 24),
	}
}

// BEARGB reads a color.ARGB x from a big endian int32.
func (r *Reader) BEARGB(x *color.RGBA) {
	var v int32
	r.BEInt32(&v)
	*x = color.RGBA{
		A: byte(v),
		R: byte(v >> 8),
		G: byte(v >> 16),
		B: byte(v >> 24),
	}
}

// VarRGBA reads a color.RGBA x from a varuint32.
func (r *Reader) VarRGBA(x *color.RGBA) {
	var v uint32
	r.Varuint32(&v)
	*x = color.RGBA{
		R: byte(v),
		G: byte(v >> 8),
		B: byte(v >> 16),
		A: byte(v >> 24),
	}
}

// Bytes reads the leftover bytes into a byte slice.
func (r *Reader) Bytes(p *[]byte) {
	var err error
	*p, err = io.ReadAll(r.r)
	if err != nil {
		r.panic(err)
	}
}

// NBT reads a compound tag into a map from the underlying buffer.
func (r *Reader) NBT(m *map[string]any, encoding nbt.Encoding) {
	dec := nbt.NewDecoderWithEncoding(r.r, encoding)
	dec.AllowZero = true

	*m = make(map[string]any)
	if err := dec.Decode(m); err != nil {
		r.panic(err)
	}
}

// NBTList reads a list of NBT tags from the underlying buffer.
func (r *Reader) NBTList(m *[]any, encoding nbt.Encoding) {
	if err := nbt.NewDecoderWithEncoding(r.r, encoding).Decode(m); err != nil {
		r.panic(err)
	}
}

// UUID reads a uuid.UUID from the underlying buffer.
func (r *Reader) UUID(x *uuid.UUID) {
	b := make([]byte, 16)
	if _, err := r.r.Read(b); err != nil {
		r.panic(err)
	}

	// The UUIDs we read are Little Endian, but the uuid library is based on Big Endian UUIDs, so we need to
	// reverse the two int64s the UUID is composed of, then reverse their bytes too.
	b = append(b[8:], b[:8]...)
	var arr [16]byte
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = b[j], b[i]
	}
	*x = arr
}

// PlayerInventoryAction reads a PlayerInventoryAction.
func (r *Reader) PlayerInventoryAction(x *UseItemTransactionData) {
	r.Varint32(&x.LegacyRequestID)
	if x.LegacyRequestID < -1 && (x.LegacyRequestID&1) == 0 {
		Slice(r, &x.LegacySetItemSlots)
	}
	Slice(r, &x.Actions)
	r.Varuint32(&x.ActionType)
	r.Varuint32(&x.TriggerType)
	r.BlockPos(&x.BlockPosition)
	r.Varint32(&x.BlockFace)
	r.Varint32(&x.HotBarSlot)
	r.ItemInstance(&x.HeldItem)
	r.Vec3(&x.Position)
	r.Vec3(&x.ClickedPosition)
	r.Varuint32(&x.BlockRuntimeID)
	r.Varuint32(&x.ClientPrediction)
}

// GameRule reads a GameRule x from the Reader.
func (r *Reader) GameRule(x *GameRule) {
	r.String(&x.Name)
	r.Bool(&x.CanBeModifiedByPlayer)
	var t uint32
	r.Varuint32(&t)

	switch t {
	case 1:
		var v bool
		r.Bool(&v)
		x.Value = v
	case 2:
		var v uint32
		r.Uint32(&v)
		x.Value = v
	case 3:
		var v float32
		r.Float32(&v)
		x.Value = v
	default:
		r.UnknownEnumOption(t, "game rule type")
	}
}

// GameRuleLegacy reads a legacy GameRule x from the Reader.
func (r *Reader) GameRuleLegacy(x *GameRule) {
	r.String(&x.Name)
	r.Bool(&x.CanBeModifiedByPlayer)
	var t uint32
	r.Varuint32(&t)

	switch t {
	case 1:
		var v bool
		r.Bool(&v)
		x.Value = v
	case 2:
		var v uint32
		r.Varuint32(&v)
		x.Value = v
	case 3:
		var v float32
		r.Float32(&v)
		x.Value = v
	default:
		r.UnknownEnumOption(t, "game rule type")
	}
}

// EntityMetadata reads an entity metadata map from the underlying buffer into map x.
func (r *Reader) EntityMetadata(x *map[uint32]any) {
	*x = map[uint32]any{}

	var count uint32
	r.Varuint32(&count)
	for i := uint32(0); i < count; i++ {
		var key, dataType uint32
		r.Varuint32(&key)
		r.Varuint32(&dataType)
		switch dataType {
		case EntityDataTypeByte:
			var v byte
			r.Uint8(&v)
			(*x)[key] = v
		case EntityDataTypeInt16:
			var v int16
			r.Int16(&v)
			(*x)[key] = v
		case EntityDataTypeInt32:
			var v int32
			r.Varint32(&v)
			(*x)[key] = v
		case EntityDataTypeFloat32:
			var v float32
			r.Float32(&v)
			(*x)[key] = v
		case EntityDataTypeString:
			var v string
			r.String(&v)
			(*x)[key] = v
		case EntityDataTypeCompoundTag:
			var v map[string]any
			r.NBT(&v, nbt.NetworkLittleEndian)
			(*x)[key] = v
		case EntityDataTypeBlockPos:
			var v BlockPos
			r.BlockPos(&v)
			(*x)[key] = v
		case EntityDataTypeInt64:
			var v int64
			r.Varint64(&v)
			(*x)[key] = v
		case EntityDataTypeVec3:
			var v mgl32.Vec3
			r.Vec3(&v)
			(*x)[key] = v
		default:
			r.UnknownEnumOption(dataType, "entity metadata")
		}
	}
}

// ItemDescriptorCount reads an ItemDescriptorCount i from the underlying buffer.
func (r *Reader) ItemDescriptorCount(i *ItemDescriptorCount) {
	var id uint8
	r.Uint8(&id)

	switch id {
	case ItemDescriptorInvalid:
		i.Descriptor = &InvalidItemDescriptor{}
	case ItemDescriptorDefault:
		i.Descriptor = &DefaultItemDescriptor{}
	case ItemDescriptorMoLang:
		i.Descriptor = &MoLangItemDescriptor{}
	case ItemDescriptorItemTag:
		i.Descriptor = &ItemTagItemDescriptor{}
	case ItemDescriptorDeferred:
		i.Descriptor = &DeferredItemDescriptor{}
	case ItemDescriptorComplexAlias:
		i.Descriptor = &ComplexAliasItemDescriptor{}
	default:
		r.UnknownEnumOption(id, "item descriptor type")
		return
	}

	i.Descriptor.Marshal(r)
	r.Varint32(&i.Count)
}

// ItemInstance reads an ItemInstance i from the underlying buffer.
func (r *Reader) ItemInstance(i *ItemInstance) {
	x := &i.Stack
	x.NBTData = make(map[string]any)
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there is no more data we should read for the item instance. After all, air
		// items aren't really anything.
		x.MetadataValue, x.Count, x.CanBePlacedOn, x.CanBreak = 0, 0, nil, nil
		return
	}

	r.Uint16(&x.Count)
	r.Varuint32(&x.MetadataValue)

	var hasNetID bool
	r.Bool(&hasNetID)

	if hasNetID {
		r.Varint32(&i.StackNetworkID)
	}

	r.Varint32(&x.BlockRuntimeID)

	var extraData []byte
	r.ByteSlice(&extraData)

	buf := bytes.NewBuffer(extraData)
	bufReader := NewReader(buf, r.shieldID, r.limitsEnabled)

	var length int16
	bufReader.Int16(&length)

	if length == -1 {
		var version uint8
		bufReader.Uint8(&version)

		switch version {
		case 1:
			bufReader.NBT(&x.NBTData, nbt.LittleEndian)
		default:
			bufReader.UnknownEnumOption(version, "item user data version")
			return
		}
	} else if length > 0 {
		bufReader.NBT(&x.NBTData, nbt.LittleEndian)
	}

	FuncSliceUint32Length(bufReader, &x.CanBePlacedOn, bufReader.StringUTF)
	FuncSliceUint32Length(bufReader, &x.CanBreak, bufReader.StringUTF)

	if x.NetworkID == bufReader.shieldID {
		var blockingTick int64
		bufReader.Int64(&blockingTick)
	}
}

// Item reads an ItemStack x from the underlying buffer.
func (r *Reader) Item(x *ItemStack) {
	x.NBTData = make(map[string]any)
	r.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there is no more data we should read for the item instance. After all, air
		// items aren't really anything.
		x.MetadataValue, x.Count, x.CanBePlacedOn, x.CanBreak = 0, 0, nil, nil
		return
	}

	r.Uint16(&x.Count)
	r.Varuint32(&x.MetadataValue)
	r.Varint32(&x.BlockRuntimeID)

	var extraData []byte
	r.ByteSlice(&extraData)

	buf := bytes.NewBuffer(extraData)
	bufReader := NewReader(buf, r.shieldID, r.limitsEnabled)

	var length int16
	bufReader.Int16(&length)

	if length == -1 {
		var version uint8
		bufReader.Uint8(&version)

		switch version {
		case 1:
			bufReader.NBT(&x.NBTData, nbt.LittleEndian)
		default:
			bufReader.UnknownEnumOption(version, "item user data version")
			return
		}
	} else if length > 0 {
		bufReader.NBT(&x.NBTData, nbt.LittleEndian)
	}

	FuncSliceUint32Length(bufReader, &x.CanBePlacedOn, bufReader.StringUTF)
	FuncSliceUint32Length(bufReader, &x.CanBreak, bufReader.StringUTF)

	if x.NetworkID == bufReader.shieldID {
		var blockingTick int64
		bufReader.Int64(&blockingTick)
	}
}

// StackRequestAction reads a StackRequestAction from the reader.
func (r *Reader) StackRequestAction(x *StackRequestAction) {
	var id uint8
	r.Uint8(&id)
	if !lookupStackRequestAction(id, x) {
		r.UnknownEnumOption(id, "stack request action type")
		return
	}
	(*x).Marshal(r)
}

// MaterialReducer reads a material reducer from the reader.
func (r *Reader) MaterialReducer(m *MaterialReducer) {
	var mix int32
	r.Varint32(&mix)
	m.InputItem = ItemType{NetworkID: mix << 16, MetadataValue: uint32(mix & 0x7fff)}
	Slice(r, &m.Outputs)
}

// Recipe reads a Recipe from the reader.
func (r *Reader) Recipe(x *Recipe) {
	var recipeType int32
	r.Varint32(&recipeType)
	if !lookupRecipe(recipeType, x) {
		r.UnknownEnumOption(recipeType, "crafting data recipe type")
		return
	}
	(*x).Unmarshal(r)
}

// EventType reads an Event's type from the reader.
func (r *Reader) EventType(x *Event) {
	var t int32
	r.Varint32(&t)
	if !lookupEvent(t, x) {
		r.UnknownEnumOption(t, "event packet event type")
	}
}

// TransactionDataType reads an InventoryTransactionData type from the reader.
func (r *Reader) TransactionDataType(x *InventoryTransactionData) {
	var transactionType uint32
	r.Varuint32(&transactionType)
	if !lookupTransactionData(transactionType, x) {
		r.UnknownEnumOption(transactionType, "inventory transaction data type")
	}
}

// AbilityValue reads an ability value from the reader.
func (r *Reader) AbilityValue(x *any) {
	valType, boolVal, floatVal := uint8(0), false, float32(0)
	r.Uint8(&valType)
	r.Bool(&boolVal)
	r.Float32(&floatVal)
	switch valType {
	case 1:
		*x = boolVal
	case 2:
		*x = floatVal
	default:
		r.InvalidValue(valType, "ability value type", "must be bool or float32")
	}
}

// Bitset reads a Bitset from the reader.
func (r *Reader) Bitset(x *Bitset, size int) {
	*x = NewBitset(size)
	for i := 0; i < size; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		} else if i+bits.Len8(b) > size {
			r.panic(errBitsetOverflow)
		}

		bi := big.NewInt(int64(b & 0x7f))
		x.int.Or(x.int, bi.Lsh(bi, uint(i)))
		if b&0x80 == 0 {
			return
		}
	}

	r.panic(errBitsetOverflow)
}

// PackSetting reads a PackSetting from the reader.
func (r *Reader) PackSetting(x *PackSetting) {
	r.String(&x.Name)
	var t uint32
	r.Varuint32(&t)
	switch t {
	case PackSettingTypeFloat:
		var v float32
		r.Float32(&v)
		x.Value = v
	case PackSettingTypeBool:
		var v bool
		r.Bool(&v)
		x.Value = v
	case PackSettingTypeString:
		var v string
		r.String(&v)
		x.Value = v
	default:
		r.UnknownEnumOption(t, "pack setting")
	}
}

// SliceLimit checks if the value passed is lower than the limit passed. If
// not, the Reader panics.
func (r *Reader) SliceLimit(value uint32, max uint32) {
	if value > max && r.limitsEnabled {
		r.panicf("slice length was too long: length of %v (max %v)", value, max)
	}
}

// ShieldID returns the shield ID provided to the reader.
func (r *Reader) ShieldID() int32 {
	return r.shieldID
}

// UnknownEnumOption panics with an unknown enum option error.
func (r *Reader) UnknownEnumOption(value any, enum string) {
	r.panicf("unknown value '%v' for enum type '%v'", value, enum)
}

// InvalidValue panics with an error indicating that the value passed is not valid for a specific field.
func (r *Reader) InvalidValue(value any, forField, reason string) {
	r.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

// errVarIntOverflow is an error set if one of the Varint methods encounters a varint that does not terminate
// after 5 or 10 bytes, depending on the data type read into.
var errVarIntOverflow = errors.New("varint overflows integer")
var errBitsetOverflow = errors.New("bitset overflows size")

// Varint64 reads up to 10 bytes from the underlying buffer into an int64.
func (r *Reader) Varint64(x *int64) {
	var ux uint64
	for i := 0; i < 70; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		ux |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			*x = int64(ux >> 1)
			if ux&1 != 0 {
				*x = ^*x
			}
			return
		}
	}
	r.panic(errVarIntOverflow)
}

// Varuint64 reads up to 10 bytes from the underlying buffer into a uint64.
func (r *Reader) Varuint64(x *uint64) {
	var v uint64
	for i := 0; i < 70; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		v |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return
		}
	}
	r.panic(errVarIntOverflow)
}

// Varint32 reads up to 5 bytes from the underlying buffer into an int32.
func (r *Reader) Varint32(x *int32) {
	var ux uint32
	for i := 0; i < 35; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		ux |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			*x = int32(ux >> 1)
			if ux&1 != 0 {
				*x = ^*x
			}
			return
		}
	}
	r.panic(errVarIntOverflow)
}

// Varuint32 reads up to 5 bytes from the underlying buffer into a uint32.
func (r *Reader) Varuint32(x *uint32) {
	var v uint32
	for i := 0; i < 35; i += 7 {
		b, err := r.r.ReadByte()
		if err != nil {
			r.panic(err)
		}

		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return
		}
	}
	r.panic(errVarIntOverflow)
}

// panicf panics with the format and values passed and assigns the error created to the Reader.
func (r *Reader) panicf(format string, a ...any) {
	panic(fmt.Errorf(format, a...))
}

// panic panics with the error passed, similarly to panicf.
func (r *Reader) panic(err error) {
	panic(err)
}
