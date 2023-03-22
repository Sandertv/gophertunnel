package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCacheStatus is sent by the client to the server at the start of the game. It is sent to let the
// server know if it supports the client-side blob cache. Clients such as Nintendo Switch do not support the
// cache, and attempting to use it anyway will fail.
type ClientCacheStatus struct {
	// Enabled specifies if the blob cache is enabled. If false, the server should not attempt to use the
	// blob cache. If true, it may do so, but it may also choose not to use it.
	Enabled bool
}

// ID ...
func (pk *ClientCacheStatus) ID() uint32 {
	return IDClientCacheStatus
}

// Marshal ...
func (pk *ClientCacheStatus) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ClientCacheStatus) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ClientCacheStatus) marshal(r protocol.IO) {
	r.Bool(&pk.Enabled)
}
