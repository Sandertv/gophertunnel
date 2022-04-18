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

// Marshal ...
func (pk *StructureTemplateDataResponse) Marshal(w *protocol.Writer) {
	w.String(&pk.StructureName)
	w.Bool(&pk.Success)
	if pk.Success {
		w.NBT(&pk.StructureTemplate, nbt.NetworkLittleEndian)
	}
	w.Uint8(&pk.ResponseType)
}

// Unmarshal ...
func (pk *StructureTemplateDataResponse) Unmarshal(r *protocol.Reader) {
	r.String(&pk.StructureName)
	r.Bool(&pk.Success)
	if pk.Success {
		r.NBT(&pk.StructureTemplate, nbt.NetworkLittleEndian)
	}
	r.Uint8(&pk.ResponseType)
}
