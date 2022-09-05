package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// DimensionData is a packet sent from the server to the client containing information about data-driven dimensions
// that the server may have registered. This packet does not seem to be sent by default, rather only being sent when
// any data-driven dimensions are registered.
type DimensionData struct {
	// Definitions contain a list of data-driven dimension definitions registered on the server.
	Definitions []protocol.DimensionDefinition
}

// ID ...
func (*DimensionData) ID() uint32 {
	return IDDimensionData
}

// Marshal ...
func (pk *DimensionData) Marshal(w *protocol.Writer) {
	protocol.Slice(w, &pk.Definitions)
}

// Unmarshal ...
func (pk *DimensionData) Unmarshal(r *protocol.Reader) {
	protocol.Slice(r, &pk.Definitions)
}
