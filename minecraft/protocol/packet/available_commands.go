package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// AvailableCommands is sent by the server to send a list of all commands that the player is able to use on
// the server. This packet holds all the arguments of each commands as well, making it possible for the client
// to provide auto-completion and command usages.
type AvailableCommands struct {
	// Commands is a list of all commands that the client should show client-side. The AvailableCommands
	// packet replaces any commands sent before. It does not only add the commands that are sent in it.
	Commands []protocol.Command
	// Constraints is a list of constraints that should be applied to certain options of enums in the commands
	// above.
	Constraints []protocol.CommandEnumConstraint
}

// ID ...
func (*AvailableCommands) ID() uint32 {
	return IDAvailableCommands
}

// Marshal ...
func (pk *AvailableCommands) Marshal(buf *bytes.Buffer) {
	values, valueIndices := pk.enumValues()
	suffixes, suffixIndices := pk.suffixes()
	enums, enumIndices := pk.enums()
	dynamicEnums, dynamicEnumIndices := pk.dynamicEnums()

	// Start by writing all enum values to the buffer.
	_ = protocol.WriteVaruint32(buf, uint32(len(values)))
	for _, value := range values {
		_ = protocol.WriteString(buf, value)
	}

	// Then all suffixes.
	_ = protocol.WriteVaruint32(buf, uint32(len(suffixes)))
	for _, suffix := range suffixes {
		_ = protocol.WriteString(buf, suffix)
	}

	// After that all actual enums, which point to enum values rather than directly writing strings.
	_ = protocol.WriteVaruint32(buf, uint32(len(enums)))
	for _, enum := range enums {
		_ = protocol.WriteString(buf, enum.Type)
		_ = protocol.WriteVaruint32(buf, uint32(len(enum.Options)))
		for _, option := range enum.Options {
			writeEnumOption(buf, option, valueIndices)
		}
	}

	// Finally we write the command data which includes all usages of the commands.
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Commands)))
	for _, command := range pk.Commands {
		_ = protocol.WriteCommandData(buf, command, enumIndices, suffixIndices, dynamicEnumIndices)
	}

	// Soft enums follow, which may be changed after sending this packet.
	_ = protocol.WriteVaruint32(buf, uint32(len(dynamicEnums)))
	for _, enum := range dynamicEnums {
		_ = protocol.WriteString(buf, enum.Type)
		_ = protocol.WriteVaruint32(buf, uint32(len(enum.Options)))
		for _, option := range enum.Options {
			_ = protocol.WriteString(buf, option)
		}
	}

	// Constraints are supposed to be here, but constraints are pointless, make no sense to be in this packet
	// and are not worth implementing.
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Constraints)))
	for _, constraint := range pk.Constraints {
		_ = protocol.WriteEnumConstraint(buf, constraint, enumIndices, valueIndices)
	}
}

// Unmarshal ...
func (pk *AvailableCommands) Unmarshal(r *protocol.Reader) {
	var count uint32

	// First we read all the enum values.
	r.Varuint32(&count)
	enumValues := make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&enumValues[i])
	}

	// Then we read all suffixes.
	r.Varuint32(&count)
	suffixes := make([]string, count)
	for i := uint32(0); i < count; i++ {
		r.String(&suffixes[i])
	}

	// After that we create all enums, which are composed of pointers to the enum values above.
	r.Varuint32(&count)
	enums := make([]protocol.CommandEnum, count)
	var optionCount uint32
	for i := uint32(0); i < count; i++ {
		r.String(&enums[i].Type)
		r.Varuint32(&optionCount)
		enums[i].Options = make([]string, optionCount)
		for j := uint32(0); j < optionCount; j++ {
			enumOption(r, &enums[i].Options[j], enumValues)
		}
	}

	// We read all the commands, which will have their enums and suffixes set automatically. We don't yet set
	// the dynamic enums as we haven't read them yet.
	r.Varuint32(&count)
	pk.Commands = make([]protocol.Command, count)
	for i := uint32(0); i < count; i++ {
		protocol.CommandData(r, &pk.Commands[i], enums, suffixes)
	}

	// We first read all soft enums of the packet.
	r.Varuint32(&count)
	softEnums := make([]protocol.CommandEnum, count)
	for i := uint32(0); i < count; i++ {
		softEnums[i].Dynamic = true
		r.String(&softEnums[i].Type)

		var optionCount uint32
		r.Varuint32(&optionCount)
		softEnums[i].Options = make([]string, optionCount)
		for j := uint32(0); j < optionCount; j++ {
			r.String(&softEnums[i].Options[j])
		}
	}

	// After we've read all soft enums, we need to match them with the values that are set in the commands
	// that we read before.
	for i, command := range pk.Commands {
		for j, overload := range command.Overloads {
			for k, param := range overload.Parameters {
				if param.Type&protocol.CommandArgSoftEnum != 0 {
					offset := param.Type & 0xffff
					r.LimitUint32(offset, uint32(len(softEnums))-1)
					pk.Commands[i].Overloads[j].Parameters[k].Enum = softEnums[offset]
				}
			}
		}
	}

	r.Varuint32(&count)
	pk.Constraints = make([]protocol.CommandEnumConstraint, count)
	for i := uint32(0); i < count; i++ {
		protocol.EnumConstraint(r, &pk.Constraints[i], enums, enumValues)
	}
}

// writeEnumOption writes an enum option to buf using the value indices passed. It is written as a
// byte/uint16/uint32 depending on the size of the value indices map.
func writeEnumOption(buf *bytes.Buffer, option string, valueIndices map[string]int) {
	l := len(valueIndices)
	switch {
	case l <= math.MaxUint8:
		_ = binary.Write(buf, binary.LittleEndian, byte(valueIndices[option]))
	case l <= math.MaxUint16:
		_ = binary.Write(buf, binary.LittleEndian, uint16(valueIndices[option]))
	default:
		_ = binary.Write(buf, binary.LittleEndian, uint32(valueIndices[option]))
	}
}

// enumOption reads an enum option from buf using the enum values passed. The option is written as a
// byte/uint16/uint32, depending on the size of the enumValues slice.
func enumOption(r *protocol.Reader, option *string, enumValues []string) {
	l := len(enumValues)
	var index uint32
	switch {
	case l <= math.MaxUint8:
		var v byte
		r.Uint8(&v)
		index = uint32(v)
	case l <= math.MaxUint16:
		var v uint16
		r.Uint16(&v)
		index = uint32(v)
	default:
		r.Uint32(&index)
	}
	r.LimitUint32(index, uint32(len(enumValues))-1)
	*option = enumValues[index]
}

// enumValues runs through all commands set to the packet and collects enum values and a map of indices
// indexed with the enum values.
func (pk *AvailableCommands) enumValues() (values []string, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range pk.Commands {
		for _, alias := range command.Aliases {
			if _, ok := indices[alias]; !ok {
				indices[alias] = len(values)
				values = append(values, alias)
			}
		}
		for _, overload := range command.Overloads {
			for _, parameter := range overload.Parameters {
				for _, option := range parameter.Enum.Options {
					if _, ok := indices[option]; !ok {
						indices[option] = len(values)
						values = append(values, option)
					}
				}
			}
		}
	}
	return
}

// suffixes runs through all commands set to the packet and collects suffixes that the parameters of the
// commands may have. It returns the suffixes and a map indexed by the suffixes.
func (pk *AvailableCommands) suffixes() (suffixes []string, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range pk.Commands {
		for _, overload := range command.Overloads {
			for _, parameter := range overload.Parameters {
				if parameter.Suffix != "" {
					if _, ok := indices[parameter.Suffix]; !ok {
						indices[parameter.Suffix] = len(suffixes)
						suffixes = append(suffixes, parameter.Suffix)
					}
				}
			}
		}
	}
	return
}

// enums runs through all commands set to the packet and collects enums that the parameters of the commands
// may have. It returns the enums and a map indexed by the enums and their offsets in the slice.
func (pk *AvailableCommands) enums() (enums []protocol.CommandEnum, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range pk.Commands {
		if len(command.Aliases) > 0 {
			aliasEnum := protocol.CommandEnum{Type: command.Name + "Aliases", Options: command.Aliases}
			indices[command.Name+"Aliases"] = len(enums)
			enums = append(enums, aliasEnum)
		}
		for _, overload := range command.Overloads {
			for _, parameter := range overload.Parameters {
				if len(parameter.Enum.Options) != 0 && !parameter.Enum.Dynamic {
					if _, ok := indices[parameter.Enum.Type]; !ok {
						indices[parameter.Enum.Type] = len(enums)
						enums = append(enums, parameter.Enum)
					}
				}
			}
		}
	}
	return
}

// dynamicEnums runs through all commands set to the packet and collects dynamic enums set as parameters of
// commands. These dynamic enums may be updated over the course of the game and are written separately.
func (pk *AvailableCommands) dynamicEnums() (enums []protocol.CommandEnum, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range pk.Commands {
		for _, overload := range command.Overloads {
			for _, parameter := range overload.Parameters {
				if parameter.Enum.Dynamic {
					if _, ok := indices[parameter.Enum.Type]; !ok {
						indices[parameter.Enum.Type] = len(enums)
						enums = append(enums, parameter.Enum)
					}
				}
			}
		}
	}
	return
}
