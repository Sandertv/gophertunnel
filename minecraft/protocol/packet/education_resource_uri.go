package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// EducationResourceURI is a packet that transmits education resource settings to all clients.
type EducationResourceURI struct {
	// Resource is the resource that is being referenced.
	Resource protocol.EducationSharedResourceURI
}

// ID ...
func (*EducationResourceURI) ID() uint32 {
	return IDEducationResourceURI
}

// Marshal ...
func (pk *EducationResourceURI) Marshal(w *protocol.Writer) {
	protocol.EducationResourceURI(w, &pk.Resource)
}

// Unmarshal ...
func (pk *EducationResourceURI) Unmarshal(r *protocol.Reader) {
	protocol.EducationResourceURI(r, &pk.Resource)
}
