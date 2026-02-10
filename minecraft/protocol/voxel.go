package protocol

// SerializableVoxelCells represents a 3D grid of voxel cell data.
type SerializableVoxelCells struct {
	// XSize is the size of the grid along the X axis.
	XSize uint8
	// YSize is the size of the grid along the Y axis.
	YSize uint8
	// ZSize is the size of the grid along the Z axis.
	ZSize uint8
	// Storage is the raw cell data stored in the grid.
	Storage []uint8
}

// Marshal encodes/decodes a SerializableVoxelCells.
func (x *SerializableVoxelCells) Marshal(r IO) {
	r.Uint8(&x.XSize)
	r.Uint8(&x.YSize)
	r.Uint8(&x.ZSize)
	FuncSlice(r, &x.Storage, r.Uint8)
}

// VoxelShapeNameEntry represents a name-to-ID mapping entry for voxel shapes.
type VoxelShapeNameEntry struct {
	// Name is the name of the voxel shape.
	Name string
	// ID is the numeric ID of the voxel shape.
	ID uint16
}

// Marshal encodes/decodes a VoxelShapeNameEntry.
func (x *VoxelShapeNameEntry) Marshal(r IO) {
	r.String(&x.Name)
	r.Uint16(&x.ID)
}

// SerializableVoxelShape represents a voxel shape with cells and coordinate axes.
type SerializableVoxelShape struct {
	// Cells is the grid of cells representing solid and empty regions.
	Cells SerializableVoxelCells
	// XCoordinates is a list of X axis coordinates for the shape.
	XCoordinates []float32
	// YCoordinates is a list of Y axis coordinates for the shape.
	YCoordinates []float32
	// ZCoordinates is a list of Z axis coordinates for the shape.
	ZCoordinates []float32
}

// Marshal encodes/decodes a SerializableVoxelShape.
func (x *SerializableVoxelShape) Marshal(r IO) {
	Single(r, &x.Cells)
	FuncSlice(r, &x.XCoordinates, r.Float32)
	FuncSlice(r, &x.YCoordinates, r.Float32)
	FuncSlice(r, &x.ZCoordinates, r.Float32)
}
