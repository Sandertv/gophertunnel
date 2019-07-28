package packet

import (
	"bytes"
	"encoding/binary"
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
func (pk *UpdateSoftEnum) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.EnumType)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Options)))
	for _, option := range pk.Options {
		_ = protocol.WriteString(buf, option)
	}
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
}

// Unmarshal ...
func (pk *UpdateSoftEnum) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := chainErr(
		protocol.String(buf, &pk.EnumType),
		protocol.Varuint32(buf, &count),
	); err != nil {
		return err
	}
	pk.Options = make([]string, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.String(buf, &pk.Options[i]); err != nil {
			return err
		}
	}

	return binary.Read(buf, binary.LittleEndian, &pk.ActionType)
}
