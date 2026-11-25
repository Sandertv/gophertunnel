package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// EditorNetwork is a packet sent from the server to the client and vise-versa to communicate editor-mode related
// information. It carries a named payload variant along with its serialised data.
type EditorNetwork struct {
	// RouteToManager ...
	RouteToManager bool
	// RawVariantName is the name of the payload variant carried by the packet.
	RawVariantName string
	// RawVariantData is the serialised data of the payload variant.
	RawVariantData string
}

// ID ...
func (*EditorNetwork) ID() uint32 {
	return IDEditorNetwork
}

func (pk *EditorNetwork) Marshal(io protocol.IO) {
	io.Bool(&pk.RouteToManager)
	io.String(&pk.RawVariantName)
	io.String(&pk.RawVariantData)
}
