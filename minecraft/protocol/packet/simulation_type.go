package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

const (
	SimulationTypeGame byte = iota
	SimulationTypeEditor
	SimulationTypeTest
	SimulationTypeInvalid
)

// SimulationType is an in-progress packet. We currently do not know the use case.
type SimulationType struct {
	// SimulationType is the simulation type selected.
	SimulationType byte
}

// ID ...
func (*SimulationType) ID() uint32 {
	return IDSimulationType
}

// Marshal ...
func (pk *SimulationType) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *SimulationType) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *SimulationType) marshal(r protocol.IO) {
	r.Uint8(&pk.SimulationType)
}
