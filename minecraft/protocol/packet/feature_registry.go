package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// FeatureRegistry is a packet used to notify the client about the world generation features the server is currently
// using. This is used in combination with the client-side world generation system introduced in v1.19.20, allowing the
// client to completely generate the chunks of the world without having to rely on the server.
type FeatureRegistry struct {
	// Features is a slice of all registered world generation features.
	Features []protocol.GenerationFeature
}

// ID ...
func (f FeatureRegistry) ID() uint32 {
	return IDFeatureRegistry
}

// Marshal ...
func (f FeatureRegistry) Marshal(w *protocol.Writer) {
	l := uint32(len(f.Features))
	w.Varuint32(&l)
	for _, feature := range f.Features {
		protocol.GenFeature(w, &feature)
	}
}

// Unmarshal ...
func (f FeatureRegistry) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.Varuint32(&count)

	f.Features = make([]protocol.GenerationFeature, count)
	for i := uint32(0); i < count; i++ {
		protocol.GenFeature(r, &f.Features[i])
	}
}
