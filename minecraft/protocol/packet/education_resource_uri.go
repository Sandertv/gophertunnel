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
	pk.marshal(w)
}

// Unmarshal ...
func (pk *EducationResourceURI) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *EducationResourceURI) marshal(r protocol.IO) {
	protocol.EducationResourceURI(r, &pk.Resource)
}
