package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	SoftEnumActionAdd = iota
	SoftEnumActionRemove
	SoftEnumActionSet
)

// UpdateSoftEnum is sent by the server to update a soft enum, also known as a dynamic enum, previously sent
// in the AvailableCommands packet. It is sent whenever the enum should get new options or when some of its
// options should be removed.
// The UpdateSoftEnum packet will apply for enums that have been set in the AvailableCommands packet with the
// 'Dynamic' field of the CommandEnum set to true.
type UpdateSoftEnum struct {
	// EnumType is the type of the enum. This type must be identical to the one set in the AvailableCommands
	// packet, because the client uses this to recognise which enum to update.
	EnumType string
	// Options is a list of options that should be updated. Depending on the ActionType field, either these
	// options will be added to the enum, the enum options will be set to these options or all of these
	// options will be removed from the enum.
	Options []string
	// ActionType is the type of the action to execute on the enum. The Options field has a different result,
	// depending on what ActionType is used.
	ActionType byte
}

// ID ...
func (*UpdateSoftEnum) ID() uint32 {
	return IDUpdateSoftEnum
}

// Marshal ...
func (pk *UpdateSoftEnum) Marshal(w *protocol.Writer) {
	w.String(&pk.EnumType)
	l := uint32(len(pk.Options))
	w.Varuint32(&l)
	for _, option := range pk.Options {
		w.String(&option)
	}
	w.Uint8(&pk.ActionType)
}

// Unmarshal ...
func (pk *UpdateSoftEnum) Unmarshal(r *protocol.Reader) {
	var count uint32
	r.String(&pk.EnumType)
	r.Varuint32(&count)

	pk.Options = make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&pk.Options[i])
	}
	r.Uint8(&pk.ActionType)
}
