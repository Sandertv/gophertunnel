package protocol

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"math/big"
	"reflect"
	"sort"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/internal"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// Writer implements writing methods for data types from Minecraft packets. Each Packet implementation has one
// passed to it when writing.
// Writer implements methods where values are passed using a pointer, so that Reader and Writer have a
// synonymous interface and both implement the IO interface.
type Writer struct {
	w interface {
		io.Writer
		io.ByteWriter
	}
	shieldID int32
	buf      [8]byte
}

// NewWriter creates a new initialised Writer with an underlying io.ByteWriter to write to.
func NewWriter(w interface {
	io.Writer
	io.ByteWriter
}, shieldID int32) *Writer {
	writer := new(Writer)
	writer.Reset(w, shieldID)
	return writer
}

// Reset reuses w with a new underlying destination and shield ID.
func (w *Writer) Reset(dst interface {
	io.Writer
	io.ByteWriter
}, shieldID int32) {
	w.w = dst
	w.shieldID = shieldID
}

// Uint8 writes a uint8 to the underlying buffer.
func (w *Writer) Uint8(x *uint8) {
	_ = w.w.WriteByte(*x)
}

// Int8 writes an int8 to the underlying buffer.
func (w *Writer) Int8(x *int8) {
	_ = w.w.WriteByte(byte(*x) & 0xff)
}

// Bool writes a bool as either 0 or 1 to the underlying buffer.
func (w *Writer) Bool(x *bool) {
	_ = w.w.WriteByte(*(*byte)(unsafe.Pointer(x)))
}

// StringUTF ...
func (w *Writer) StringUTF(x *string) {
	l := int16(len(*x))
	w.Int16(&l)
	w.writeString(*x)
}

// String writes a string, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) String(x *string) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	w.writeString(*x)
}

// stringWriter is implemented by writers that support writing strings directly.
type stringWriter interface {
	WriteString(string) (int, error)
}

// writeString uses WriteString when available to avoid converting s to []byte.
func (w *Writer) writeString(s string) {
	if sw, ok := w.w.(stringWriter); ok {
		_, _ = sw.WriteString(s)
		return
	}
	_, _ = w.w.Write([]byte(s))
}

// ByteSlice writes a []byte, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) ByteSlice(x *[]byte) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	_, _ = w.w.Write(*x)
}

// Bytes appends a []byte to the underlying buffer.
func (w *Writer) Bytes(x *[]byte) {
	_, _ = w.w.Write(*x)
}

// ByteFloat writes a rotational float32 as a single byte to the underlying buffer.
func (w *Writer) ByteFloat(x *float32) {
	_ = w.w.WriteByte(byte(*x / (360.0 / 256.0)))
}

// Vec3 writes an mgl32.Vec3 as 3 float32s to the underlying buffer.
func (w *Writer) Vec3(x *mgl32.Vec3) {
	w.Float32(&x[0])
	w.Float32(&x[1])
	w.Float32(&x[2])
}

// Vec2 writes an mgl32.Vec2 as 2 float32s to the underlying buffer.
func (w *Writer) Vec2(x *mgl32.Vec2) {
	w.Float32(&x[0])
	w.Float32(&x[1])
}

// BlockPos writes a BlockPos as 3 varint32s to the underlying buffer.
func (w *Writer) BlockPos(x *BlockPos) {
	w.Varint32(&x[0])
	w.Varint32(&x[1])
	w.Varint32(&x[2])
}

// ChunkPos writes a ChunkPos as 2 varint32s to the underlying buffer.
func (w *Writer) ChunkPos(x *ChunkPos) {
	w.Varint32(&x[0])
	w.Varint32(&x[1])
}

// SubChunkPos writes a SubChunkPos as 3 varint32s to the underlying buffer.
func (w *Writer) SubChunkPos(x *SubChunkPos) {
	w.Varint32(&x[0])
	w.Varint32(&x[1])
	w.Varint32(&x[2])
}

// SoundPos writes an mgl32.Vec3 that serves as a position for a sound.
func (w *Writer) SoundPos(x *mgl32.Vec3) {
	b := BlockPos{int32((*x)[0] * 8), int32((*x)[1] * 8), int32((*x)[2] * 8)}
	w.BlockPos(&b)
}

// RGB writes a color.RGBA x as 3 float32s to the underlying buffer.
func (w *Writer) RGB(x *color.RGBA) {
	red := float32(x.R) / 255
	green := float32(x.G) / 255
	blue := float32(x.B) / 255
	w.Float32(&red)
	w.Float32(&green)
	w.Float32(&blue)
}

// RGBA writes a color.RGBA x as a uint32 to the underlying buffer.
func (w *Writer) RGBA(x *color.RGBA) {
	val := uint32(x.R) | uint32(x.G)<<8 | uint32(x.B)<<16 | uint32(x.A)<<24
	w.Uint32(&val)
}

// ARGB writes a color.RGBA x as a int32 to the underlying buffer.
func (w *Writer) ARGB(x *color.RGBA) {
	val := int32(x.A) | int32(x.R)<<8 | int32(x.G)<<16 | int32(x.B)<<24
	w.Int32(&val)
}

// BEARGB writes a color.RGBA x as a big endian int32 to the underlying buffer.
func (w *Writer) BEARGB(x *color.RGBA) {
	val := int32(x.A) | int32(x.R)<<8 | int32(x.G)<<16 | int32(x.B)<<24
	w.BEInt32(&val)
}

// VarRGBA writes a color.RGBA x as a varuint32 to the underlying buffer.
func (w *Writer) VarRGBA(x *color.RGBA) {
	val := uint32(x.R) | uint32(x.G)<<8 | uint32(x.B)<<16 | uint32(x.A)<<24
	w.Varuint32(&val)
}

// UUID writes a UUID to the underlying buffer.
func (w *Writer) UUID(x *uuid.UUID) {
	b := append((*x)[8:], (*x)[:8]...)
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	_, _ = w.w.Write(b)
}

// PlayerInventoryAction writes a PlayerInventoryAction.
func (w *Writer) PlayerInventoryAction(x *UseItemTransactionData) {
	w.Varint32(&x.LegacyRequestID)
	legacy := x.LegacyRequestID < -1 && (x.LegacyRequestID&1) == 0
	w.Bool(&legacy)
	if legacy {
		Slice(w, &x.LegacySetItemSlots)
	}
	// The actions are wrapped in a struct that is itself optional, so their
	// presence is stated by two booleans rather than one.
	present := true
	w.Bool(&present)
	w.Bool(&present)
	Slice(w, &x.Actions)
	// Everything from the action type onwards is identical to the transaction
	// data carried by an InventoryTransaction packet.
	x.Marshal(w)
}

// GameRule writes a GameRule x to the Writer.
func (w *Writer) GameRule(x *GameRule) {
	w.String(&x.Name)
	w.Bool(&x.CanBeModifiedByPlayer)

	switch v := x.Value.(type) {
	case nil:
		id := uint32(0)
		w.Varuint32(&id)
	case bool:
		id := uint32(1)
		w.Varuint32(&id)
		w.Bool(&v)
	case uint32:
		id := uint32(2)
		w.Varuint32(&id)
		w.Uint32(&v)
	case float32:
		id := uint32(3)
		w.Varuint32(&id)
		w.Float32(&v)
	default:
		w.UnknownEnumOption(fmt.Sprintf("%T", v), "game rule type")
	}
}

// EntityMetadata writes an entity metadata map x to the underlying buffer.
func (w *Writer) EntityMetadata(x *EntityMetadata) {
	l := uint32(len(*x))
	w.Varuint32(&l)

	// Entity metadata needs to be sorted for some functionality to work. NPCs, for example, need to have their fields
	// set in increasing order, or the text or buttons won't be shown to the client. See #88.
	// Sorting this is probably not very fast, but it'll have to do for now: We can change entity metadata to a slice
	// later on.
	keys := make([]int, 0, l)
	for k := range *x {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		key := uint32(k)
		value := (*x)[uint32(k)]
		w.Varuint32(&key)
		switch v := value.(type) {
		case byte:
			w.entityDataType(EntityDataTypeByte)
			w.Uint8(&v)
		case int16:
			w.entityDataType(EntityDataTypeInt16)
			w.Int16(&v)
		case int32:
			w.entityDataType(EntityDataTypeInt32)
			w.Varint32(&v)
		case float32:
			w.entityDataType(EntityDataTypeFloat32)
			w.Float32(&v)
		case string:
			w.entityDataType(EntityDataTypeString)
			w.String(&v)
		case map[string]any:
			w.entityDataType(EntityDataTypeCompoundTag)
			w.NBT(&v, nbt.NetworkLittleEndian)
		case BlockPos:
			w.entityDataType(EntityDataTypeBlockPos)
			w.BlockPos(&v)
		case int64:
			w.entityDataType(EntityDataTypeInt64)
			w.Varint64(&v)
		case mgl32.Vec3:
			w.entityDataType(EntityDataTypeVec3)
			w.Vec3(&v)
		default:
			w.UnknownEnumOption(reflect.TypeOf(value), "entity metadata")
		}
	}
}

// entityDataType writes the type of an entity metadata value. The type is sent
// twice: first as a varuint32 and then as a byte.
func (w *Writer) entityDataType(t uint32) {
	b := byte(t)
	w.Varuint32(&t)
	w.Uint8(&b)
}

// ItemDescriptorCount writes an ItemDescriptorCount i to the underlying buffer.
func (w *Writer) ItemDescriptorCount(i *ItemDescriptorCount) {
	var name string
	if !lookupItemDescriptorName(i.Descriptor, &name) {
		w.UnknownEnumOption(fmt.Sprintf("%T", i.Descriptor), "item descriptor type")
		return
	}
	present := uint32(1)
	if name == "" {
		present = 0
	}
	w.Varuint32(&present)

	descriptor := i.Descriptor
	if name == "" {
		descriptor = &InvalidItemDescriptor{}
	} else {
		w.String(&name)
	}

	descriptor.Marshal(w)
	w.Varint32(&i.Count)
}

// ItemInstance writes an ItemInstance i to the underlying buffer.
func (w *Writer) ItemInstance(i *ItemInstance) {
	x := &i.Stack
	id := int16(x.NetworkID)
	w.Int16(&id)

	w.Uint16(&x.Count)
	w.Varuint32(&x.MetadataValue)

	hasNetID := i.StackNetworkID != 0
	w.Bool(&hasNetID)

	if hasNetID {
		w.Varint32(&i.StackNetworkID)
	}

	runtimeID := uint32(x.BlockRuntimeID)
	w.Varuint32(&runtimeID)

	if x.NetworkID == 0 {
		var zero uint32
		w.Varuint32(&zero)
		return
	}
	w.itemUserData(&x.NBTData, &x.CanBePlacedOn, &x.CanBreak, &x.BlockingTick, x.NetworkID == w.shieldID)
}

// Item writes an ItemStack x to the underlying buffer.
func (w *Writer) Item(x *ItemStack) {
	w.Varint32(&x.NetworkID)

	w.Uint16(&x.Count)
	w.Varuint32(&x.MetadataValue)
	w.Varint32(&x.BlockRuntimeID)

	if x.NetworkID == 0 {
		// Air items carry an empty user data buffer, matching the server.
		var zero uint32
		w.Varuint32(&zero)
		return
	}
	w.itemUserData(&x.NBTData, &x.CanBePlacedOn, &x.CanBreak, &x.BlockingTick, x.NetworkID == w.shieldID)
}

// StackReqItemDescriptorCount writes an ItemDescriptorCount i as it appears in an ItemStackRequest. This
// encoding differs from the one used by recipes: the type is a numeric variant index repeated as a byte
// rather than a name, and the count is a little endian uint16.
func (w *Writer) StackReqItemDescriptorCount(i *ItemDescriptorCount) {
	var t uint8
	if !lookupItemDescriptorType(i.Descriptor, &t) {
		w.UnknownEnumOption(fmt.Sprintf("%T", i.Descriptor), "item descriptor type")
		return
	}
	control := uint32(t)
	w.Varuint32(&control)
	w.Uint8(&t)

	// The payload of a descriptor here is not the one a recipe ingredient uses: neither the invalid nor the
	// item tag descriptor trails a metadata value.
	switch d := i.Descriptor.(type) {
	case *DeferredItemDescriptor:
		w.String(&d.Name)
		IntegerFunc(&d.MetadataValue, w.Varint32)
	case *MoLangItemDescriptor:
		w.String(&d.Expression)
		w.Int16(&d.Version)
	case *ItemTagItemDescriptor:
		w.String(&d.Tag)
	}

	count := uint16(i.Count)
	w.Uint16(&count)
}

// StackReqItem writes a StackRequestItem x to the writer.
func (w *Writer) StackReqItem(x *StackRequestItem) {
	t := ItemDescriptorTypeItemName
	if x.Name == "" {
		t = ItemDescriptorTypeInvalid
	}
	control := uint32(t)
	w.Varuint32(&control)
	w.Uint8(&t)
	if t != ItemDescriptorTypeInvalid {
		w.String(&x.Name)
		w.Varint32(&x.MetadataValue)
	}
	w.Int16(&x.Count)
	w.Varuint32(&x.BlockRuntimeID)
	w.itemUserData(&x.NBTData, &x.CanBePlacedOn, &x.CanBreak, &x.BlockingTick, x.Name == ItemNameShield)
}

// itemUserData writes the user data buffer that trails an item, which holds its NBT and the blocks it may be
// placed on or break. blocking is set for the shield only, which trails the tick it started blocking at.
func (w *Writer) itemUserData(nbtData *map[string]any, canBePlacedOn, canBreak *[]string, blockingTick *int64, blocking bool) {
	buf := internal.BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		buf.Reset()
		internal.BufferPool.Put(buf)
	}()
	bufWriter := NewWriter(buf, w.shieldID)

	var length int16
	if len(*nbtData) != 0 {
		length = int16(-1)
		version := uint8(1)

		bufWriter.Int16(&length)
		bufWriter.Uint8(&version)
		bufWriter.NBT(nbtData, nbt.LittleEndian)
	} else {
		bufWriter.Int16(&length)
	}

	FuncSliceUint32Length(bufWriter, canBePlacedOn, bufWriter.StringUTF)
	FuncSliceUint32Length(bufWriter, canBreak, bufWriter.StringUTF)

	if blocking {
		bufWriter.Int64(blockingTick)
	}

	extraData := buf.Bytes()
	w.ByteSlice(&extraData)
}

// StackRequestAction writes a StackRequestAction to the writer.
func (w *Writer) StackRequestAction(x *StackRequestAction) {
	var id byte
	if !lookupStackRequestActionType(*x, &id) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "stack request action type")
	}
	// The variant is selected by its index in the variant list, and the type is
	// then repeated as the variant's own first field.
	control := stackRequestActionVariant(id)
	w.Varuint32(&control)
	w.Uint8(&id)
	(*x).Marshal(w)
}

// MaterialReducer writes a material reducer to the writer.
func (w *Writer) MaterialReducer(m *MaterialReducer) {
	mix := (m.InputItem.NetworkID << 16) | int32(m.InputItem.MetadataValue)
	w.Varint32(&mix)
	Slice(w, &m.Outputs)
}

// EventType writes an Event to the writer.
func (w *Writer) EventType(x *Event) {
	var t int32
	if !lookupEventType(*x, &t) {
		w.UnknownEnumOption(*x, "event packet event type")
	}
	w.Varint32(&t)
}

// EventOrdinal writes an Event ordinal to the writer.
func (w *Writer) EventOrdinal(x *Event) {
	var ordinal uint32
	if !lookupEventOrdinal(*x, &ordinal) {
		w.UnknownEnumOption(*x, "event packet event ordinal")
		return
	}
	w.Varuint32(&ordinal)
}

// TransactionDataType writes an InventoryTransactionData type to the writer.
func (w *Writer) TransactionDataType(x *InventoryTransactionData) {
	var id uint32
	if !lookupTransactionDataType(*x, &id) {
		w.UnknownEnumOption(fmt.Sprintf("%T", x), "inventory transaction data type")
	}
	w.Varuint32(&id)
}

// AbilityValue writes an ability value to the writer.
func (w *Writer) AbilityValue(x *any) {
	switch val := (*x).(type) {
	case bool:
		valType, defaultVal := uint8(1), float32(0)
		w.Uint8(&valType)
		w.Bool(&val)
		w.Float32(&defaultVal)
	case float32:
		valType, defaultVal := uint8(2), false
		w.Uint8(&valType)
		w.Bool(&defaultVal)
		w.Float32(&val)
	default:
		w.InvalidValue(*x, "ability value type", "must be bool or float32")
	}
}

var varintMaxByteValue = big.NewInt(0x80)

// Bitset writes a Bitset x to the underlying buffer.
func (w *Writer) Bitset(x *Bitset, size int) {
	if x.size != size {
		w.panicf("bitset size mismatch: expected %v, got %v", size, x.size)
	}
	u := new(big.Int)
	u.Set(x.int)

	if len(u.Bits()) == 0 {
		_ = w.w.WriteByte(0)
		return
	}

	for u.Cmp(varintMaxByteValue) >= 0 {
		_ = w.w.WriteByte(byte(u.Bits()[0]) | 0x80)
		u.Rsh(u, 7)
	}
	_ = w.w.WriteByte(byte(u.Bits()[0]))
}

// PackSetting writes a PackSetting x to the underlying buffer.
func (w *Writer) PackSetting(x *PackSetting) {
	w.String(&x.Name)
	var id uint32
	switch val := x.Value.(type) {
	case float32:
		id = PackSettingTypeFloat
		w.Varuint32(&id)
		w.Float32(&val)
	case bool:
		id = PackSettingTypeBool
		w.Varuint32(&id)
		w.Bool(&val)
	case string:
		id = PackSettingTypeString
		w.Varuint32(&id)
		w.String(&val)
	default:
		w.UnknownEnumOption(x.Value, "pack setting")
	}
}

// ShapeData writes a ShapeData to the writer.
func (w *Writer) ShapeData(x *ShapeData) {
	var shapeDataType uint32
	if !lookupShapeDataType(*x, &shapeDataType) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "debug shape data type")
	}
	w.Varuint32(&shapeDataType)
	(*x).Marshal(w)
}

// Varint64 writes an int64 as 1-10 bytes to the underlying buffer.
func (w *Writer) Varint64(x *int64) {
	u := *x
	ux := uint64(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		_ = w.w.WriteByte(byte(ux) | 0x80)
		ux >>= 7
	}
	_ = w.w.WriteByte(byte(ux))
}

// Varuint64 writes a uint64 as 1-10 bytes to the underlying buffer.
func (w *Writer) Varuint64(x *uint64) {
	u := *x
	for u >= 0x80 {
		_ = w.w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	_ = w.w.WriteByte(byte(u))
}

// Varint32 writes an int32 as 1-5 bytes to the underlying buffer.
func (w *Writer) Varint32(x *int32) {
	u := *x
	ux := uint32(u) << 1
	if u < 0 {
		ux = ^ux
	}
	for ux >= 0x80 {
		_ = w.w.WriteByte(byte(ux) | 0x80)
		ux >>= 7
	}
	_ = w.w.WriteByte(byte(ux))
}

// Varuint32 writes a uint32 as 1-5 bytes to the underlying buffer.
func (w *Writer) Varuint32(x *uint32) {
	u := *x
	for u >= 0x80 {
		_ = w.w.WriteByte(byte(u) | 0x80)
		u >>= 7
	}
	_ = w.w.WriteByte(byte(u))
}

// NBT writes a map as NBT to the underlying buffer using the encoding passed.
func (w *Writer) NBT(x *map[string]any, encoding nbt.Encoding) {
	if err := nbt.NewEncoderWithEncoding(w.w, encoding).Encode(*x); err != nil {
		panic(err)
	}
}

// NBTList writes a slice as NBT to the underlying buffer using the encoding passed.
func (w *Writer) NBTList(x *[]any, encoding nbt.Encoding) {
	if err := nbt.NewEncoderWithEncoding(w.w, encoding).Encode(*x); err != nil {
		panic(err)
	}
}

// ShieldID returns the shield ID provided to the writer.
func (w *Writer) ShieldID() int32 {
	return w.shieldID
}

// UnknownEnumOption panics with an unknown enum option error.
func (w *Writer) UnknownEnumOption(value any, enum string) {
	w.panicf("unknown value '%#v' for enum type '%v'", value, enum)
}

// InvalidValue panics with an invalid value error.
func (w *Writer) InvalidValue(value any, forField, reason string) {
	w.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

// panicf panics with the format and values passed.
func (w *Writer) panicf(format string, a ...any) {
	panic(fmt.Errorf(format, a...))
}
