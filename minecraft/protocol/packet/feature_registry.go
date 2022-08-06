package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// FeatureRegistry is a packet used to notify the client about the world generation features the server is currently
// using. This is used in combination with the client-side world generation system introduced in v1.19.20, allowing the
// client to completely generate the chunks of the world without having to rely on the server.
type FeatureRegistry struct {
	// Features is a slice of all registered world generation features.
	Features []protocol.GenerationFeature
}

// ID ...
func (pk *FeatureRegistry) ID() uint32 {
	return IDFeatureRegistry
}

// Marshal ...
func (pk *FeatureRegistry) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Features))
	w.Varuint32(&l)
	for _, feature := range pk.Features {
		protocol.GenFeature(w, &feature)
	}
}

// Unmarshal ...
func (pk *FeatureRegistry) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)

	pk.Features = make([]protocol.GenerationFeature, count)
	for i := uint32(0); i < count; i++ {
		protocol.GenFeature(r, &pk.Features[i])
	}
}
