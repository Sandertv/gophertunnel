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
	// CustomDataTagName is the name of a function to run, usually used during natural generation. A
	// description can be found here: https://minecraft.gamepedia.com/Structure_Block#Data.
	CustomDataTagName string
	// IncludePlayers specifies if the 'Include Players' toggle has been enabled, meaning players are also
	// exported by the structure block.
	IncludePlayers bool
	// DetectStructureSizeAndPosition specifies if the size and position of the selection of the structure
	// block should be detected using another structure block. Currently, this field is inaccessible in the
	// game. It is likely to be added in v1.13.
	DetectStructureSizeAndPosition bool
	// StructureBlockType is the type of the structure block updated. In Bedrock Edition v1.12, this will
	// always be '5': The 3D Export structure block. According to the Minecraft wiki, v1.13 adds other types
	// of structure blocks to the game, so expect changes once that happens.
	// A list of structure block types that will be used can be found in the constants above.
	StructureBlockType int32
	// Settings is a struct of settings that should be used for exporting the structure. These settings are
	// identical to the last sent in the StructureBlockUpdate packet by the client.
	Settings protocol.StructureSettings
	// Bool1 ...
	Bool1 bool
}

// ID ...
func (*StructureBlockUpdate) ID() uint32 {
	return IDStructureBlockUpdate
}

// Marshal ...
func (pk *StructureBlockUpdate) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = protocol.WriteString(buf, pk.StructureName)
	_ = protocol.WriteString(buf, pk.CustomDataTagName)
	_ = binary.Write(buf, binary.LittleEndian, pk.IncludePlayers)
	_ = binary.Write(buf, binary.LittleEndian, pk.DetectStructureSizeAndPosition)
	_ = protocol.WriteVarint32(buf, pk.StructureBlockType)
	_ = protocol.WriteStructSettings(buf, pk.Settings)
	_ = binary.Write(buf, binary.LittleEndian, pk.Bool1)
}

// Unmarshal ...
func (pk *StructureBlockUpdate) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.UBlockPosition(buf, &pk.Position),
		protocol.String(buf, &pk.StructureName),
		protocol.String(buf, &pk.CustomDataTagName),
		binary.Read(buf, binary.LittleEndian, &pk.IncludePlayers),
		binary.Read(buf, binary.LittleEndian, &pk.DetectStructureSizeAndPosition),
		protocol.Varint32(buf, &pk.StructureBlockType),
		protocol.StructSettings(buf, &pk.Settings),
		binary.Read(buf, binary.LittleEndian, &pk.Bool1),
	)
}
