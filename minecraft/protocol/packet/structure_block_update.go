package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	StructureBlockData = iota
	StructureBlockSave
	StructureBlockLoad
	StructureBlockCorner
	StructureBlockInvalid
	StructureBlockExport
)

const (
	StructureRedstoneSaveModeMemory = iota
	StructureRedstoneSaveModeDisk
)

// StructureBlockUpdate is sent by the client when it updates a structure block using the in-game UI. The
// data it contains depends on the type of structure block that it is. In Minecraft Bedrock Edition v1.11,
// there is only the Export structure block type, but in v1.13 the ones present in Java Edition will,
// according to the wiki, be added too.
type StructureBlockUpdate struct {
	// Position is the position of the structure block that is updated.
	Position protocol.BlockPos
	// StructureName is the name of the structure that was set in the structure block's UI. This is the name
	// used to export the structure to a file.
	StructureName string
	// FilteredStructureName is a filtered version of StructureName with all the profanity removed. The client
	// will use this over StructureName if this field is not empty and they have the "Filter Profanity"
	// setting enabled.
	FilteredStructureName string
	// DataField is the name of a function to run, usually used during natural generation. A description can
	// be found here: https://minecraft.wiki/w/Structure_Block#Data.
	DataField string
	// IncludePlayers specifies if the 'Include Players' toggle has been enabled, meaning players are also
	// exported by the structure block.
	IncludePlayers bool
	// ShowBoundingBox specifies if the structure block should have its bounds outlined. A thin line will
	// encapsulate the bounds of the structure if set to true.
	ShowBoundingBox bool
	// StructureBlockType is the type of the structure block updated. A list of structure block types that
	// will be used can be found in the constants above.
	StructureBlockType int32
	// Settings is a struct of settings that should be used for exporting the structure. These settings are
	// identical to the last sent in the StructureBlockUpdate packet by the client.
	Settings protocol.StructureSettings
	// RedstoneSaveMode is the mode that should be used to save the structure when used with redstone. In
	// Java Edition, this is always stored in memory, but in Bedrock Edition it can be stored either to disk
	// or memory. See the constants above for the options.
	RedstoneSaveMode int32
	// ShouldTrigger specifies if the structure block should be triggered immediately after this packet
	// reaches the server.
	ShouldTrigger bool
	// Waterlogged specifies if non-air blocks replace water or combine with water.
	Waterlogged bool
}

// ID ...
func (*StructureBlockUpdate) ID() uint32 {
	return IDStructureBlockUpdate
}

func (pk *StructureBlockUpdate) Marshal(io protocol.IO) {
	io.UBlockPos(&pk.Position)
	io.String(&pk.StructureName)
	io.String(&pk.FilteredStructureName)
	io.String(&pk.DataField)
	io.Bool(&pk.IncludePlayers)
	io.Bool(&pk.ShowBoundingBox)
	io.Varint32(&pk.StructureBlockType)
	protocol.Single(io, &pk.Settings)
	io.Varint32(&pk.RedstoneSaveMode)
	io.Bool(&pk.ShouldTrigger)
	io.Bool(&pk.Waterlogged)
}
