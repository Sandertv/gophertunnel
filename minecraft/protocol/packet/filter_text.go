package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// FilterText ...Cus
type FilterText struct {
	// Text ...
	Text string
	// FromServer ...
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
