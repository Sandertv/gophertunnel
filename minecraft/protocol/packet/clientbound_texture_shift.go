package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientBoundTextureShift is sent by the server to control texture shift animations on the client.
type ClientBoundTextureShift struct {
	// ActionID is the texture shift action to perform. It is one of the TextureShiftAction constants.
	ActionID uint8
	// CollectionName is the name of the texture shift collection.
	CollectionName string
	// FromStep is the step to shift from.
	FromStep string
	// ToStep is the step to shift to.
	ToStep string
	// AllSteps is a list of all steps in the texture shift.
	AllSteps []string
	// CurrentLengthTicks is the current length of the shift in ticks.
	CurrentLengthTicks uint64
	// TotalLengthTicks is the total length of the shift in ticks.
	TotalLengthTicks uint64
	// Enabled specifies if the texture shift is enabled.
	Enabled bool
}

// ID ...
func (*ClientBoundTextureShift) ID() uint32 {
	return IDClientBoundTextureShift
}

func (pk *ClientBoundTextureShift) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionID)
	io.String(&pk.CollectionName)
	io.String(&pk.FromStep)
	io.String(&pk.ToStep)
	protocol.FuncSlice(io, &pk.AllSteps, io.String)
	io.Varuint64(&pk.CurrentLengthTicks)
	io.Varuint64(&pk.TotalLengthTicks)
	io.Bool(&pk.Enabled)
}
