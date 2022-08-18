package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ServerStats ...
// TODO: Document this.
type ServerStats struct {
	// ServerTime ...
	ServerTime float32
	// NetworkTime ...
	NetworkTime float32
}

// ID ...
func (pk *ServerStats) ID() uint32 {
	return IDServerStats
}

// Marshal ...
func (pk *ServerStats) Marshal(w *protocol.Writer) {
	w.Float32(&pk.ServerTime)
	w.Float32(&pk.NetworkTime)
}

// Unmarshal ...
func (pk *ServerStats) Unmarshal(r *protocol.Reader) {
	r.Float32(&pk.ServerTime)
	r.Float32(&pk.NetworkTime)
}
