package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// StructureTemplateDataRequest is sent by the client to request data of a structure. It is currently unused:
// The client never sends this packet.
type StructureTemplateDataRequest struct {
	// StructureName is the name of the structure that was set in the structure block's UI. This is the name
	// used to export the structure to a file.
	StructureName string
	// Position is the position of the structure block that has its template data requested.
	Position protocol.BlockPos
	// Settings is a struct of settings that should be used for exporting the structure. These settings are
	// identical to the last sent in the StructureBlockUpdate packet by the client.
	Settings protocol.StructureSettings
	// ContainerID ...
	Byte1 byte
}

// ID ...
func (pk *StructureTemplateDataRequest) ID() uint32 {
	return IDStructureTemplateDataRequest
}

// Marshal ...
func (pk *StructureTemplateDataRequest) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.StructureName)
	_ = protocol.WriteUBlockPosition(buf, pk.Position)
	_ = protocol.WriteStructSettings(buf, pk.Settings)
	_ = binary.Write(buf, binary.LittleEndian, pk.Byte1)
}

// Unmarshal ...
func (pk *StructureTemplateDataRequest) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.StructureName),
		protocol.UBlockPosition(buf, &pk.Position),
		protocol.StructSettings(buf, &pk.Settings),
		binary.Read(buf, binary.LittleEndian, &pk.Byte1),
	)
}
