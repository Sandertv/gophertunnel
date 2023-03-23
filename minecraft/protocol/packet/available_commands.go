package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
func (pk *AvailableCommands) Marshal(w *protocol.Writer) {
	values, valueIndices := pk.enumValues()
	suffixes, suffixIndices := pk.suffixes()
	enums, enumIndices := pk.enums()
	dynamicEnums, dynamicEnumIndices := pk.dynamicEnums()

	ctx := protocol.AvailableCommandsContext{
		EnumIndices:        enumIndices,
		EnumValueIndices:   valueIndices,
		SuffixIndices:      suffixIndices,
		DynamicEnumIndices: dynamicEnumIndices,
	}

	// Start by writing all enum values and suffixes to the buffer.
	protocol.FuncSlice(w, &values, w.String)
	protocol.FuncSlice(w, &suffixes, w.String)

	// After that all actual enums, which point to enum values rather than directly writing strings.
	protocol.FuncIOSlice(w, &enums, ctx.WriteEnum)

	// Finally we write the command data which includes all usages of the commands.
	protocol.FuncIOSlice(w, &pk.Commands, ctx.WriteCommandData)

	// Soft enums follow, which may be changed after sending this packet.
	protocol.Slice(w, &dynamicEnums)

	protocol.FuncIOSlice(w, &pk.Constraints, ctx.WriteEnumConstraint)
}

// Unmarshal ...
func (pk *AvailableCommands) Unmarshal(r *protocol.Reader) {
	var ctx protocol.AvailableCommandsContext

	// First we read all the enum values and suffixes.
	protocol.FuncSlice(r, &ctx.EnumValues, r.String)
	protocol.FuncSlice(r, &ctx.Suffixes, r.String)

	// After that we create all enums, which are composed of pointers to the enum values above.
	protocol.FuncIOSlice(r, &ctx.Enums, ctx.Enum)

	// We read all the commands, which will have their enums and suffixes set automatically. We don't yet set
	// the dynamic enums as we haven't read them yet.
	protocol.FuncIOSlice(r, &pk.Commands, ctx.CommandData)

	// We first read all soft enums of the packet.
	protocol.Slice(r, &ctx.DynamicEnums)

	protocol.FuncIOSlice(r, &pk.Constraints, ctx.EnumConstraint)

	// After we've read all soft enums, we need to match them with the values that are set in the commands
	// that we read before.
	for i, command := range pk.Commands {
		for j, overload := range command.Overloads {
			for k, param := range overload.Parameters {
				if param.Type&protocol.CommandArgSoftEnum != 0 {
					pk.Commands[i].Overloads[j].Parameters[k].Enum = ctx.DynamicEnums[param.Type&0xffff]
				}
			}
		}
	}
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
