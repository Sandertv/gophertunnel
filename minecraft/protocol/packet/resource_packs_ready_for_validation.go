package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ResourcePacksReadyForValidation is sent by the client to inform the server that the client has finished
// loading resource packs and is ready for validation.
type ResourcePacksReadyForValidation struct{}

// ID ...
func (*ResourcePacksReadyForValidation) ID() uint32 {
	return IDResourcePacksReadyForValidation
}

func (pk *ResourcePacksReadyForValidation) Marshal(protocol.IO) {}
