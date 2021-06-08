package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

const (
	StructureMirrorNone = iota
	StructureMirrorXAxis
	StructureMirrorZAxis
	StructureMirrorBothAxes
)

const (
	StructureRotationNone = iota
	StructureRotationRotate90
	StructureRotationRotate180
	StructureRotationRotate270
)

const (
	AnimationModeNone = iota
	AnimationModeLayers
	AnimationModeBlocks
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
	// LastEditingPlayerUniqueID is the unique ID of the player that last edited the structure block that
	// these settings concern.
	LastEditingPlayerUniqueID int64
	// Rotation is the rotation that the structure block should obtain. See the constants above for available
	// options.
	Rotation byte
	// Mirror specifies the way the structure should be mirrored. It is either no mirror at all, mirror on the
	// x/z axis or both.
	Mirror byte
	// AnimationMode ...
	AnimationMode byte
	// AnimationDuration ...
	AnimationDuration float32
	// Integrity is usually 1, but may be set to a number between 0 and 1 to omit blocks randomly, using
	// the Seed that follows.
	Integrity float32
	// Seed is the seed used to omit blocks if Integrity is not equal to one. If the Seed is 0, a random
	// seed is selected to omit blocks.
	Seed uint32
	// Pivot is the pivot around which the structure may be rotated.
	Pivot mgl32.Vec3
}

// StructSettings reads/writes StructureSettings x using IO r.
func StructSettings(r IO, x *StructureSettings) {
	r.String(&x.PaletteName)
	r.Bool(&x.IgnoreEntities)
	r.Bool(&x.IgnoreBlocks)
	r.UBlockPos(&x.Size)
	r.UBlockPos(&x.Offset)
	r.Varint64(&x.LastEditingPlayerUniqueID)
	r.Uint8(&x.Rotation)
	r.Uint8(&x.Mirror)
	r.Uint8(&x.AnimationMode)
	r.Float32(&x.AnimationDuration)
	r.Float32(&x.Integrity)
	r.Uint32(&x.Seed)
	r.Vec3(&x.Pivot)
}
