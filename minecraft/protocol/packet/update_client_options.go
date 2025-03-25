package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	GraphicsModeSimple = iota
	GraphicsModeFancy
	GraphicsModeAdvanced
	GraphicsModeRayTraced
)

// UpdateClientOptions is sent by the client when some of the client's options are updated, such as the
// graphics mode.
type UpdateClientOptions struct {
	// GraphicsMode is the graphics mode that the client is using. It is one of the constants above.
	GraphicsMode protocol.Optional[byte]
}

// ID ...
func (*UpdateClientOptions) ID() uint32 {
	return IDUpdateClientOptions
}

func (pk *UpdateClientOptions) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.GraphicsMode, io.Uint8)
}
