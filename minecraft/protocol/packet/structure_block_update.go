package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	StructureBlockData = iota + 1
	StructureBlockSave
	StructureBlockLoad
	StructureBlockCorner
	StructureBlockExport
)

const (
	StructureMirrorNone = iota
	StructureMirrorLeftToRight
	StructureMirrorFrontToBack
)

// StructureBlockUpdate is sent by the client when it updates a structure block using the in-game UI. The
// data it contains depends on the type of structure block that it is. In Minecraft Bedrock Edition v1.11,
// there is only the Export structure block type, but in v1.13 the ones present in Java Edition will,
// according to the wiki, be added too.
type StructureBlockUpdate struct {
	// Position is the position of the structure block that is updated.
	Position protocol.BlockPos
	// StructureBlockType is the type of the structure block updated. In Bedrock Edition v1.11, this will
	// always be '5': The 3D Export structure block. According to the Minecraft wiki, v1.13 adds other types
	// of structure blocks to the game, so expect changes once that happens.
	// A list of structure block types that will be used can be found in the constants above.
	StructureBlockType uint32

	// StructureName is the name of the structure that was set in the structure block's UI. This is the name
	// used to export the structure to a file.
	StructureName string
	// CustomDataTagName is the name of a function to run, usually used during natural generation. A
	// description can be found here: https://minecraft.gamepedia.com/Structure_Block#Data.
	CustomDataTagName string
	// Offset is the offset position that was set in the structure block. The area exported is offset by this
	// position.
	Offset protocol.BlockPos
	// Size is the size of the area that is about to be exported. The area exported will start at the
	// Position + Offset, and will extend as far as Size specifies.
	Size protocol.BlockPos
	// ExcludeEntities specifies if entities should be excluded from the structure block export. It seems
	// rather counter-intuitive that this field is not called 'IncludeEntities', but it holds a value that
	// implies the exclusion of entities.
	ExcludeEntities bool
	// RemoveBlocks specifies if the 'Remove Blocks' toggle has been enabled, meaning that no blocks will be
	// exported from the structure block.
	RemoveBlocks bool
	// IncludePlayers specifies if the 'Include Players' toggle has been enabled, meaning players are also
	// exported by the structure block.
	IncludePlayers bool
	// DetectStructureSizeAndPosition specifies if the size and position of the selection of the structure
	// block should be detected using another structure block. Currently, this field is inaccessible in the
	// game. It is likely to be added in v1.13.
	DetectStructureSizeAndPosition bool

	// LegacyData is a struct holding data that is not currently used by Bedrock Edition structure blocks. It
	// appears to be in use in v1.13.
	LegacyData struct {
		// Integrity is usually 1, but may be set to a number between 0 and 1 to omit blocks randomly, using
		// the Seed that follows.
		Integrity float32
		// Seed is the seed used to omit blocks if Integrity is not equal to one. If the Seed is 0, a random
		// seed is selected to omit blocks.
		Seed uint32
		// Mirror specifies the way the structure should be mirrored. It is either no mirror at all, left to
		// right mirror or front to back, as the constants above specify.
		Mirror uint32
		// Rotation is the rotation that the structure block should obtain.
		Rotation uint32
		// IgnoreStructureBlocks specifies if the structure blocks within the selection of this structure
		// block should be ignored.
		IgnoreStructureBlocks bool
		// BoundingBoxCornerOne is used if StructureBlockType is StructureBlockCorner. If not, it does not
		// represent an actual position: It always holds [2147483647 -2 2147483647].
		BoundingBoxCornerOne protocol.BlockPos
		// BoundingBoxCornerOne is used if StructureBlockType is StructureBlockCorner. If not, it does not
		// represent an actual position: It always holds [-2147483647 -3 -2147483647].
		BoundingBoxCornerTwo protocol.BlockPos
	}

	// Functionality of these fields is unknown.
	UnknownBool1 bool
	UnknownBool2 bool
}

// ID ...
func (*StructureBlockUpdate) ID() uint32 {
	return IDStructureBlockUpdate
}

// Marshal ...
func (pk *StructureBlockUpdate) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = protocol.WriteVaruint32(buf, pk.StructureBlockType)
	_ = protocol.WriteString(buf, pk.StructureName)
	_ = protocol.WriteString(buf, pk.CustomDataTagName)
	_ = protocol.WriteUBlockPosition(buf, pk.Offset)
	_ = protocol.WriteUBlockPosition(buf, pk.Size)
	_ = binary.Write(buf, binary.LittleEndian, pk.ExcludeEntities)
	_ = binary.Write(buf, binary.LittleEndian, pk.RemoveBlocks)
	_ = binary.Write(buf, binary.LittleEndian, pk.IncludePlayers)
	_ = binary.Write(buf, binary.LittleEndian, pk.DetectStructureSizeAndPosition)

	_ = protocol.WriteFloat32(buf, pk.LegacyData.Integrity)
	_ = protocol.WriteVaruint32(buf, pk.LegacyData.Seed)
	_ = protocol.WriteVaruint32(buf, pk.LegacyData.Mirror)
	_ = protocol.WriteVaruint32(buf, pk.LegacyData.Rotation)
	_ = binary.Write(buf, binary.LittleEndian, pk.LegacyData.IgnoreStructureBlocks)
	_ = protocol.WriteUBlockPosition(buf, pk.LegacyData.BoundingBoxCornerOne)
	_ = protocol.WriteUBlockPosition(buf, pk.LegacyData.BoundingBoxCornerTwo)

	_ = binary.Write(buf, binary.LittleEndian, pk.UnknownBool1)
	_ = binary.Write(buf, binary.LittleEndian, pk.UnknownBool2)
}

// Unmarshal ...
func (pk *StructureBlockUpdate) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.UBlockPosition(buf, &pk.Position),
		protocol.Varuint32(buf, &pk.StructureBlockType),
		protocol.String(buf, &pk.StructureName),
		protocol.String(buf, &pk.CustomDataTagName),
		protocol.UBlockPosition(buf, &pk.Offset),
		protocol.UBlockPosition(buf, &pk.Size),
		binary.Read(buf, binary.LittleEndian, &pk.ExcludeEntities),
		binary.Read(buf, binary.LittleEndian, &pk.RemoveBlocks),
		binary.Read(buf, binary.LittleEndian, &pk.IncludePlayers),
		binary.Read(buf, binary.LittleEndian, &pk.DetectStructureSizeAndPosition),

		protocol.Float32(buf, &pk.LegacyData.Integrity),
		protocol.Varuint32(buf, &pk.LegacyData.Seed),
		protocol.Varuint32(buf, &pk.LegacyData.Mirror),
		protocol.Varuint32(buf, &pk.LegacyData.Rotation),
		binary.Read(buf, binary.LittleEndian, &pk.LegacyData.IgnoreStructureBlocks),
		protocol.UBlockPosition(buf, &pk.LegacyData.BoundingBoxCornerOne),
		protocol.UBlockPosition(buf, &pk.LegacyData.BoundingBoxCornerTwo),

		binary.Read(buf, binary.LittleEndian, &pk.UnknownBool1),
		binary.Read(buf, binary.LittleEndian, &pk.UnknownBool2),
	)
}
