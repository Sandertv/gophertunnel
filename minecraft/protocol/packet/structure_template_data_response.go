package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	StructureTemplateResponseExport = iota + 1
	StructureTemplateResponseQuery
)

// StructureTemplateDataResponse is sent by the server to send data of a structure to the client in response
// to a StructureTemplateDataRequest packet.
type StructureTemplateDataResponse struct {
	// StructureName is the name of the structure that was requested. This is the name used to export the
	// structure to a file.
	StructureName string
	// Success specifies if a structure template was found by the StructureName that was sent in a
	// StructureTemplateDataRequest packet.
	Success bool
	// ResponseType specifies the response type of the packet. This depends on the RequestType field sent in
	// the StructureTemplateDataRequest packet and is one of the constants above.
	ResponseType byte
	// StructureTemplate holds the data of the structure template.
	StructureTemplate map[string]interface{}
}

// ID ...
func (pk *StructureTemplateDataResponse) ID() uint32 {
	return IDStructureTemplateDataResponse
}

// Marshal ...
func (pk *StructureTemplateDataResponse) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.StructureName)
	_ = binary.Write(buf, binary.LittleEndian, pk.Success)
	if pk.Success {
		if err := nbt.NewEncoder(buf).Encode(pk.StructureTemplate); err != nil {
			panic(err)
		}
	}
	_ = binary.Write(buf, binary.LittleEndian, pk.ResponseType)
}

// Unmarshal ...
func (pk *StructureTemplateDataResponse) Unmarshal(buf *bytes.Buffer) error {
	var success bool
	if err := chainErr(
		protocol.String(buf, &pk.StructureName),
		binary.Read(buf, binary.LittleEndian, &success),
	); err != nil {
		return err
	}
	if success {
		if err := nbt.NewDecoder(buf).Decode(&pk.StructureTemplate); err != nil {
			return err
		}
	}
	return binary.Read(buf, binary.LittleEndian, &pk.ResponseType)
}
