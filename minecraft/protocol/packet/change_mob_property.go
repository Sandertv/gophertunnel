package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ChangeMobProperty is a packet sent from the server to the client to change one of the properties of a mob client-side.
type ChangeMobProperty struct {
	// EntityUniqueID is the unique ID of the entity whose property is being changed.
	EntityUniqueID uint64
	// Property is the name of the property being updated.
	Property string
	// BoolValue is set if the property value is a bool type. If the type is not a bool, this field is ignored.
	BoolValue bool
	// StringValue is set if the property value is a string type. If the type is not a string, this field is ignored.
	StringValue string
	// IntValue is set if the property value is an int type. If the type is not an int, this field is ignored.
	IntValue int32
	// FloatValue is set if the property value is a float type. If the type is not a float, this field is ignored.
	FloatValue float32
}

// ID ...
func (*ChangeMobProperty) ID() uint32 {
	return IDChangeMobProperty
}

// Marshal ...
func (pk *ChangeMobProperty) Marshal(w *protocol.Writer) {
	w.Uint64(&pk.EntityUniqueID)
	w.String(&pk.Property)
	w.Bool(&pk.BoolValue)
	w.String(&pk.StringValue)
	w.Varint32(&pk.IntValue)
	w.Float32(&pk.FloatValue)
}

// Unmarshal ...
func (pk *ChangeMobProperty) Unmarshal(r *protocol.Reader) {
	r.Uint64(&pk.EntityUniqueID)
	r.String(&pk.Property)
	r.Bool(&pk.BoolValue)
	r.String(&pk.StringValue)
	r.Varint32(&pk.IntValue)
	r.Float32(&pk.FloatValue)
}
