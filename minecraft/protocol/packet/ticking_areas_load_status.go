package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// TickingAreasLoadStatus is sent by the server to the client to notify the client of a ticking area's loading status.
type TickingAreasLoadStatus struct {
	// Preload is true if the server is waiting for the area's preload.
	Preload bool
}

// ID ...
func (*TickingAreasLoadStatus) ID() uint32 {
	return IDTickingAreasLoadStatus
}

// Marshal ...
func (pk *TickingAreasLoadStatus) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *TickingAreasLoadStatus) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *TickingAreasLoadStatus) marshal(r protocol.IO) {
	r.Bool(&pk.Preload)
}
