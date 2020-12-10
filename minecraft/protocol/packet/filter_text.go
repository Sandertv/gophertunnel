package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// FilterText is sent by the both the client and the server. The client sends the packet to the server to
// allow the server to filter the text server-side. The server then responds with the same packet and the
// safer version of the text.
type FilterText struct {
	// Text is either the text from the client or the safer version of the text sent by the server.
	Text string
	// FromServer indicates if the packet was sent by the server or not.
	FromServer bool
}

// ID ...
func (*FilterText) ID() uint32 {
	return IDFilterText
}

// Marshal ...
func (pk *FilterText) Marshal(w *protocol.Writer) {
	w.String(&pk.Text)
	w.Bool(&pk.FromServer)
}

// Unmarshal ...
func (pk *FilterText) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Text)
	r.Bool(&pk.FromServer)
}
