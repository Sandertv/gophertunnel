package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AnimateEntity is sent by the server to animate an entity client-side.
type AnimateEntity struct {
	// TODO: Document all these fields.
	Animation     string
	NextState     string
	StopCondition string
	Controller    string
	BlendOutTime  float32
	// EntityRuntimeIDs is list of runtime IDs of entities that the animation should be applied to.
	EntityRuntimeIDs []uint64
}

// ID ...
func (*AnimateEntity) ID() uint32 {
	return IDAnimateEntity
}

// Marshal ...
func (pk *AnimateEntity) Marshal(w *protocol.Writer) {
	w.String(&pk.Animation)
	w.String(&pk.NextState)
	w.String(&pk.StopCondition)
	w.String(&pk.Controller)
	w.Float32(&pk.BlendOutTime)
	l := uint32(len(pk.EntityRuntimeIDs))
	w.Varuint32(&l)
	for i := range pk.EntityRuntimeIDs {
		w.Varuint64(&pk.EntityRuntimeIDs[i])
	}
}

// Unmarshal ...
func (pk *AnimateEntity) Unmarshal(r *protocol.Reader) {
	r.String(&pk.Animation)
	r.String(&pk.NextState)
	r.String(&pk.StopCondition)
	r.String(&pk.Controller)
	var count uint32
	r.Varuint32(&count)
	pk.EntityRuntimeIDs = make([]uint64, count)
	for i := uint32(0); i < count; i++ {
		r.Varuint64(&pk.EntityRuntimeIDs[i])
	}
}
