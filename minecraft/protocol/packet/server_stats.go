package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ServerStats is a packet sent from the server to the client to update the client on server statistics. It is purely
// used for telemetry.
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
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ServerStats) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ServerStats) marshal(r protocol.IO) {
	r.Float32(&pk.ServerTime)
	r.Float32(&pk.NetworkTime)
}
