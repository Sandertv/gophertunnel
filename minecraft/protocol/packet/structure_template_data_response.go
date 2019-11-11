package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// StructureTemplateDataExportResponse is sent by the server to send data of a structure to the client in
// response to a StructureTemplateDataRequest packet. It is currently unused: The former packet is not
// sent by the client.
type StructureTemplateDataExportResponse struct {
	// StructureName is the name of the structure that was requested. This is the name used to export the
	// structure to a file.
	StructureName string
	// SerialisedStructureTemplate is a network little endian NBT serialised structure of the structure
	// template. It holds all the data of the structure.
	SerialisedStructureTemplate []byte
}

// ID ...
func (pk *StructureTemplateDataExportResponse) ID() uint32 {
	return IDStructureTemplateDataResponse
}

// Marshal ...
func (pk *StructureTemplateDataExportResponse) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.StructureName)
	_ = binary.Write(buf, binary.LittleEndian, len(pk.SerialisedStructureTemplate) != 0)
	_, _ = buf.Write(pk.SerialisedStructureTemplate)
}

// Unmarshal ...
func (pk *StructureTemplateDataExportResponse) Unmarshal(buf *bytes.Buffer) error {
	var hasData bool
	if err := chainErr(
		protocol.String(buf, &pk.StructureName),
		binary.Read(buf, binary.LittleEndian, &hasData),
	); err != nil {
		return err
	}
	if hasData {
		pk.SerialisedStructureTemplate = buf.Next(math.MaxInt32)
	}
	return nil
}
