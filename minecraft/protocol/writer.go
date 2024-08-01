package protocol

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"image/color"
	"io"
	"reflect"
	"sort"
	"unsafe"
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
}

// NewWriter creates a new initialised Writer with an underlying io.ByteWriter to write to.
func NewWriter(w interface {
	io.Writer
	io.ByteWriter
}, shieldID int32) *Writer {
	return &Writer{w: w, shieldID: shieldID}
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
	_, _ = w.w.Write([]byte(*x))
}

// String writes a string, prefixed with a varuint32, to the underlying buffer.
func (w *Writer) String(x *string) {
	l := uint32(len(*x))
	w.Varuint32(&l)
	_, _ = w.w.Write([]byte(*x))
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

// UBlockPos writes a BlockPos as 2 varint32s and a varuint32 to the underlying buffer.
func (w *Writer) UBlockPos(x *BlockPos) {
	w.Varint32(&x[0])
	y := uint32(x[1])
	w.Varuint32(&y)
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
	w.UBlockPos(&b)
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
	if x.LegacyRequestID < -1 && (x.LegacyRequestID&1) == 0 {
		Slice(w, &x.LegacySetItemSlots)
	}
	Slice(w, &x.Actions)
	w.Varuint32(&x.ActionType)
	w.BlockPos(&x.BlockPosition)
	w.Varint32(&x.BlockFace)
	w.Varint32(&x.HotBarSlot)
	w.ItemInstance(&x.HeldItem)
	w.Vec3(&x.Position)
	w.Vec3(&x.ClickedPosition)
	w.Varuint32(&x.BlockRuntimeID)
}

// GameRule writes a GameRule x to the Writer.
func (w *Writer) GameRule(x *GameRule) {
	w.String(&x.Name)
	w.Bool(&x.CanBeModifiedByPlayer)

	switch v := x.Value.(type) {
	case bool:
		id := uint32(1)
		w.Varuint32(&id)
		w.Bool(&v)
	case uint32:
		id := uint32(2)
		w.Varuint32(&id)
		w.Varuint32(&v)
	case float32:
		id := uint32(3)
		w.Varuint32(&id)
		w.Float32(&v)
	default:
		w.UnknownEnumOption(fmt.Sprintf("%T", v), "game rule type")
	}
}

// EntityMetadata writes an entity metadata map x to the underlying buffer.
func (w *Writer) EntityMetadata(x *map[uint32]any) {
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
			entityDataTypeByte := EntityDataTypeByte
			w.Varuint32(&entityDataTypeByte)
			w.Uint8(&v)
		case int16:
			entityDataTypeInt16 := EntityDataTypeInt16
			w.Varuint32(&entityDataTypeInt16)
			w.Int16(&v)
		case int32:
			entityDataTypeInt32 := EntityDataTypeInt32
			w.Varuint32(&entityDataTypeInt32)
			w.Varint32(&v)
		case float32:
			entityDataTypeFloat32 := EntityDataTypeFloat32
			w.Varuint32(&entityDataTypeFloat32)
			w.Float32(&v)
		case string:
			entityDataTypeString := EntityDataTypeString
			w.Varuint32(&entityDataTypeString)
			w.String(&v)
		case map[string]any:
			entityDataTypeCompoundTag := EntityDataTypeCompoundTag
			w.Varuint32(&entityDataTypeCompoundTag)
			w.NBT(&v, nbt.NetworkLittleEndian)
		case BlockPos:
			entityDataTypeBlockPos := EntityDataTypeBlockPos
			w.Varuint32(&entityDataTypeBlockPos)
			w.BlockPos(&v)
		case int64:
			entityDataTypeInt64 := EntityDataTypeInt64
			w.Varuint32(&entityDataTypeInt64)
			w.Varint64(&v)
		case mgl32.Vec3:
			entityDataTypeVec3 := EntityDataTypeVec3
			w.Varuint32(&entityDataTypeVec3)
			w.Vec3(&v)
		default:
			w.UnknownEnumOption(reflect.TypeOf(value), "entity metadata")
		}
	}
}

// ItemDescriptorCount writes an ItemDescriptorCount i to the underlying buffer.
func (w *Writer) ItemDescriptorCount(i *ItemDescriptorCount) {
	var id byte
	switch i.Descriptor.(type) {
	case *InvalidItemDescriptor:
		id = ItemDescriptorInvalid
	case *DefaultItemDescriptor:
		id = ItemDescriptorDefault
	case *MoLangItemDescriptor:
		id = ItemDescriptorMoLang
	case *ItemTagItemDescriptor:
		id = ItemDescriptorItemTag
	case *DeferredItemDescriptor:
		id = ItemDescriptorDeferred
	case *ComplexAliasItemDescriptor:
		id = ItemDescriptorComplexAlias
	default:
		w.UnknownEnumOption(fmt.Sprintf("%T", i.Descriptor), "item descriptor type")
		return
	}
	w.Uint8(&id)

	i.Descriptor.Marshal(w)
	w.Varint32(&i.Count)
}

// ItemInstance writes an ItemInstance i to the underlying buffer.
func (w *Writer) ItemInstance(i *ItemInstance) {
	x := &i.Stack
	w.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there's no more data to follow. Return immediately.
		return
	}

	w.Uint16(&x.Count)
	w.Varuint32(&x.MetadataValue)

	hasNetID := i.StackNetworkID != 0
	w.Bool(&hasNetID)

	if hasNetID {
		w.Varint32(&i.StackNetworkID)
	}

	w.Varint32(&x.BlockRuntimeID)

	buf := new(bytes.Buffer)
	bufWriter := NewWriter(buf, w.shieldID)

	var length int16
	if len(x.NBTData) != 0 {
		length = int16(-1)
		version := uint8(1)

		bufWriter.Int16(&length)
		bufWriter.Uint8(&version)
		bufWriter.NBT(&x.NBTData, nbt.LittleEndian)
	} else {
		bufWriter.Int16(&length)
	}

	FuncSliceUint32Length(bufWriter, &x.CanBePlacedOn, bufWriter.StringUTF)
	FuncSliceUint32Length(bufWriter, &x.CanBreak, bufWriter.StringUTF)

	if x.NetworkID == bufWriter.shieldID {
		var blockingTick int64
		bufWriter.Int64(&blockingTick)
	}

	b := buf.Bytes()
	w.ByteSlice(&b)
}

// Item writes an ItemStack x to the underlying buffer.
func (w *Writer) Item(x *ItemStack) {
	w.Varint32(&x.NetworkID)
	if x.NetworkID == 0 {
		// The item was air, so there's no more data to follow. Return immediately.
		return
	}

	w.Uint16(&x.Count)
	w.Varuint32(&x.MetadataValue)
	w.Varint32(&x.BlockRuntimeID)

	var extraData []byte
	buf := bytes.NewBuffer(extraData)
	bufWriter := NewWriter(buf, w.shieldID)

	var length int16
	if len(x.NBTData) != 0 {
		length = int16(-1)
		version := uint8(1)

		bufWriter.Int16(&length)
		bufWriter.Uint8(&version)
		bufWriter.NBT(&x.NBTData, nbt.LittleEndian)
	} else {
		bufWriter.Int16(&length)
	}

	FuncSliceUint32Length(bufWriter, &x.CanBePlacedOn, bufWriter.StringUTF)
	FuncSliceUint32Length(bufWriter, &x.CanBreak, bufWriter.StringUTF)

	if x.NetworkID == bufWriter.shieldID {
		var blockingTick int64
		bufWriter.Int64(&blockingTick)
	}

	extraData = buf.Bytes()
	w.ByteSlice(&extraData)
}

// StackRequestAction writes a StackRequestAction to the writer.
func (w *Writer) StackRequestAction(x *StackRequestAction) {
	var id byte
	if !lookupStackRequestActionType(*x, &id) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "stack request action type")
	}
	w.Uint8(&id)
	(*x).Marshal(w)
}

// MaterialReducer writes a material reducer to the writer.
func (w *Writer) MaterialReducer(m *MaterialReducer) {
	mix := (m.InputItem.NetworkID << 16) | int32(m.InputItem.MetadataValue)
	w.Varint32(&mix)
	Slice(w, &m.Outputs)
}

// Recipe writes a Recipe to the writer.
func (w *Writer) Recipe(x *Recipe) {
	var recipeType int32
	if !lookupRecipeType(*x, &recipeType) {
		w.UnknownEnumOption(fmt.Sprintf("%T", *x), "crafting recipe type")
	}
	w.Varint32(&recipeType)
	(*x).Marshal(w)
}

// EventType writes an Event to the writer.
func (w *Writer) EventType(x *Event) {
	var t int32
	if !lookupEventType(*x, &t) {
		w.UnknownEnumOption(fmt.Sprintf("%T", x), "event packet event type")
	}
	w.Varint32(&t)
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

// CompressedBiomeDefinitions reads a list of compressed biome definitions from the reader. Minecraft decided to make their
// own type of compression for this, so we have to implement it ourselves. It uses a dictionary of repeated byte sequences
// to reduce the size of the data. The compressed data is read byte-by-byte, and if the byte is 0xff then it is assumed
// that the next two bytes are an int16 for the dictionary index. Otherwise, the byte is copied to the output. The dictionary
// index is then used to look up the byte sequence to be appended to the output.
func (w *Writer) CompressedBiomeDefinitions(x *map[string]any) {
	decompressed, err := nbt.Marshal(x)
	if err != nil {
		w.panicf("error marshaling nbt: %v", err)
	}

	var compressed []byte
	buf := bytes.NewBuffer(compressed)
	bufWriter := NewWriter(buf, w.shieldID)

	header := []byte("COMPRESSED")
	bufWriter.Bytes(&header)

	// TODO: Dictionary compression implementation
	var dictionaryLength uint16
	bufWriter.Uint16(&dictionaryLength)
	for _, b := range decompressed {
		bufWriter.Uint8(&b)
		if b == 0xff {
			dictionaryIndex := int16(1)
			bufWriter.Int16(&dictionaryIndex)
		}
	}

	compressed = buf.Bytes()
	length := uint32(len(compressed))
	w.Varuint32(&length)
	w.Bytes(&compressed)
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
	w.panicf("unknown value '%v' for enum type '%v'", value, enum)
}

// InvalidValue panics with an invalid value error.
func (w *Writer) InvalidValue(value any, forField, reason string) {
	w.panicf("invalid value '%v' for %v: %v", value, forField, reason)
}

// panicf panics with the format and values passed.
func (w *Writer) panicf(format string, a ...any) {
	panic(fmt.Errorf(format, a...))
}
