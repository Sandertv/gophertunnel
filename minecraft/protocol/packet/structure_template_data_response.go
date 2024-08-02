package packet

import (
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
	StructureTemplate map[string]any
}

// ID ...
func (pk *StructureTemplateDataResponse) ID() uint32 {
	return IDStructureTemplateDataResponse
}

func (pk *StructureTemplateDataResponse) Marshal(io protocol.IO) {
	io.String(&pk.StructureName)
	io.Bool(&pk.Success)
	if pk.Success {
		io.NBT(&pk.StructureTemplate, nbt.NetworkLittleEndian)
	}
	io.Uint8(&pk.ResponseType)
}
