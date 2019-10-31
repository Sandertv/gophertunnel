package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	_ = protocol.WriteVaruint32(buf, 0)
}

// Unmarshal ...
func (pk *AvailableCommands) Unmarshal(buf *bytes.Buffer) error {
	var count uint32

	// First we read all the enum values.
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	enumValues := make([]string, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.String(buf, &enumValues[i]); err != nil {
			return err
		}
	}

	// Then we read all suffixes.
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	suffixes := make([]string, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.String(buf, &suffixes[i]); err != nil {
			return err
		}
	}

	// After that we create all enums, which are composed of pointers to the enum values above.
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	enums := make([]protocol.CommandEnum, count)
	var optionCount uint32
	for i := uint32(0); i < count; i++ {
		if err := protocol.String(buf, &enums[i].Type); err != nil {
			return err
		}
		if err := protocol.Varuint32(buf, &optionCount); err != nil {
			return err
		}
		enums[i].Options = make([]string, optionCount)
		for j := uint32(0); j < optionCount; j++ {
			if err := enumOption(buf, &enums[i].Options[j], enumValues); err != nil {
				return err
			}
		}
	}

	// We read all the commands, which will have their enums and suffixes set automatically. We don't yet set
	// the dynamic enums as we haven't read them yet.
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	pk.Commands = make([]protocol.Command, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.CommandData(buf, &pk.Commands[i], enums, suffixes); err != nil {
			return err
		}
	}

	// We first read all soft enums of the packet.
	if err := protocol.Varuint32(buf, &count); err != nil {
		return err
	}
	softEnums := make([]protocol.CommandEnum, count)
	for i := uint32(0); i < count; i++ {
		softEnums[i].Dynamic = true
		if err := protocol.String(buf, &softEnums[i].Type); err != nil {
			return err
		}

		var optionCount uint32
		if err := protocol.Varuint32(buf, &optionCount); err != nil {
			return err
		}
		softEnums[i].Options = make([]string, optionCount)
		for j := uint32(0); j < optionCount; j++ {
			if err := protocol.String(buf, &softEnums[i].Options[j]); err != nil {
				return err
			}
		}
	}

	// After we've read all soft enums, we need to match them with the values that are set in the commands
	// that we read before.
	for i, command := range pk.Commands {
		for j, overload := range command.Overloads {
			for k, param := range overload.Parameters {
				if param.Type&protocol.CommandArgSoftEnum != 0 {
					offset := param.Type & 0xffff
					if len(softEnums) <= int(offset) {
						return fmt.Errorf("invalid soft enum offset %v, expected lower than or equal to %v", offset, len(softEnums))
					}
					pk.Commands[i].Overloads[j].Parameters[k].Enum = softEnums[offset]
				}
			}
		}
	}

	// The constraints follow: They are useless and nonsensical, so we don't implement them.
	var enumValueSymbol, enumSymbol, constraintIndexCount uint32
	var constraintIndex byte
	_ = protocol.Varuint32(buf, &count)
	for i := uint32(0); i < count; i++ {
		_ = protocol.Varuint32(buf, &enumValueSymbol)
		_ = protocol.Varuint32(buf, &enumSymbol)
		_ = protocol.Varuint32(buf, &constraintIndexCount)
		for j := uint32(0); j < constraintIndexCount; j++ {
			_ = binary.Read(buf, binary.LittleEndian, &constraintIndex)
		}
	}

	return nil
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
func enumOption(buf *bytes.Buffer, option *string, enumValues []string) error {
	l := len(enumValues)
	var index int
	switch {
	case l <= math.MaxUint8:
		var v byte
		if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
			return err
		}
		index = int(v)
	case l <= math.MaxUint16:
		var v uint16
		if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
			return err
		}
		index = int(v)
	default:
		var v uint32
		if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
			return err
		}
		index = int(v)
	}
	if len(enumValues) <= index {
		return fmt.Errorf("invalid enum option index %v, expected lower than or equal to %v", index, len(enumValues))
	}
	*option = enumValues[index]
	return nil
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
