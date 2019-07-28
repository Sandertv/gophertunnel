package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	StructureMirrorNone = iota
	StructureMirrorLeftToRight
	StructureMirrorFrontToBack
)

// StructureSettings is a struct holding settings of a structure block. Its fields may be changed using the
// in-game UI on the client-side.
type StructureSettings struct {
	// PaletteName is the name of the palette used in the structure. Currently, it seems that this field is
	// always 'default'.
	PaletteName string
	// IgnoreEntities specifies if the structure should ignore entities or include them. If set to false,
	// entities will also show up in the exported structure.
	IgnoreEntities bool
	// IgnoreBlocks specifies if the structure should ignore blocks or include them. If set to false, blocks
	// will show up in the exported structure.
	IgnoreBlocks bool

	// Size is the size of the area that is about to be exported. The area exported will start at the
	// Position + Offset, and will extend as far as Size specifies.
	Size BlockPos
	// Offset is the offset position that was set in the structure block. The area exported is offset by this
	// position.
	Offset BlockPos
	// LastTouchedByPlayerUniqueID is the unique ID of the player that last touched the structure block that
	// these settings concern.
	LastTouchedByPlayerUniqueID int64

	// Rotation is the rotation that the structure block should obtain.
	Rotation byte
	// Mirror specifies the way the structure should be mirrored. It is either no mirror at all, left to
	// right mirror or front to back, as the constants above specify.
	Mirror byte
	// Integrity is usually 1, but may be set to a number between 0 and 1 to omit blocks randomly, using
	// the Seed that follows.
	Integrity float32
	// Seed is the seed used to omit blocks if Integrity is not equal to one. If the Seed is 0, a random
	// seed is selected to omit blocks.
	Seed uint32
}

// StructSettings reads StructureSettings x from Buffer src.
func StructSettings(src *bytes.Buffer, x *StructureSettings) error {
	return chainErr(
		String(src, &x.PaletteName),
		binary.Read(src, binary.LittleEndian, &x.IgnoreEntities),
		binary.Read(src, binary.LittleEndian, &x.IgnoreBlocks),
		UBlockPosition(src, &x.Size),
		UBlockPosition(src, &x.Offset),
		Varint64(src, &x.LastTouchedByPlayerUniqueID),
		binary.Read(src, binary.LittleEndian, &x.Rotation),
		binary.Read(src, binary.LittleEndian, &x.Mirror),
		Float32(src, &x.Integrity),
		binary.Read(src, binary.LittleEndian, &x.Seed),
	)
}

// WriteStructSettings writes StructureSettings x to Buffer dst.
func WriteStructSettings(dst *bytes.Buffer, x StructureSettings) error {
	return chainErr(
		WriteString(dst, x.PaletteName),
		binary.Write(dst, binary.LittleEndian, x.IgnoreEntities),
		binary.Write(dst, binary.LittleEndian, x.IgnoreBlocks),
		WriteUBlockPosition(dst, x.Size),
		WriteUBlockPosition(dst, x.Offset),
		WriteVarint64(dst, x.LastTouchedByPlayerUniqueID),
		binary.Write(dst, binary.LittleEndian, x.Rotation),
		binary.Write(dst, binary.LittleEndian, x.Mirror),
		WriteFloat32(dst, x.Integrity),
		binary.Write(dst, binary.LittleEndian, x.Seed),
	)
}
