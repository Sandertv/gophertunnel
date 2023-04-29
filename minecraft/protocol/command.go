package protocol

import (
	"github.com/google/uuid"
	"math"
)

// AvailableCommandsContext holds the context required to encode/decode a
// packet.AvailableCommands.
type AvailableCommandsContext struct {
	// Fields for encoding.
	EnumIndices        map[string]int
	EnumValueIndices   map[string]int
	SuffixIndices      map[string]int
	DynamicEnumIndices map[string]int

	// Fields for decoding.
	Enums        []CommandEnum
	DynamicEnums []CommandEnum
	EnumValues   []string
	Suffixes     []string
}

// Command holds the data that a command requires to be shown to a player client-side. The command is shown in
// the /help command and auto-completed using this data.
type Command struct {
	// Name is the name of the command. The command may be executed using this name, and will be shown in the
	// /help list with it. It currently seems that the client crashes if the Name contains uppercase letters.
	Name string
	// Description is the description of the command. It is shown in the /help list and when starting to write
	// a command.
	Description string
	// Flags is a combination of flags not currently known. Leaving the Flags field empty appears to work.
	Flags uint16
	// PermissionLevel is the command permission level that the player required to execute this command. The
	// field no longer seems to serve a purpose, as the client does not handle the execution of commands
	// anymore: The permissions should be checked server-side.
	PermissionLevel byte
	// Aliases is a list of aliases of the command name, that may alternatively be used to execute the command
	// in the same way.
	Aliases []string
	// Overloads is a list of command overloads that specify the ways in which a command may be executed. The
	// overloads may be completely different.
	Overloads []CommandOverload
}

// WriteCommandData writes a Command x to Writer w, using the enum indices and suffix indices passed to
// translate enums and suffixes to the indices that they're written in in the buffer.
func (ctx AvailableCommandsContext) WriteCommandData(w IO, x *Command) {
	alias := int32(-1)
	if len(x.Aliases) != 0 {
		alias = int32(ctx.EnumIndices[x.Name+"Aliases"])
	}
	w.String(&x.Name)
	w.String(&x.Description)
	w.Uint16(&x.Flags)
	w.Uint8(&x.PermissionLevel)
	w.Int32(&alias)
	FuncIOSlice(w, &x.Overloads, ctx.WriteCommandOverload)
}

// CommandData reads a Command x from Buffer src using the enums and suffixes passed to match indices with
// the values these slices hold.
func (ctx AvailableCommandsContext) CommandData(r IO, x *Command) {
	var aliasOffset int32
	r.String(&x.Name)
	r.String(&x.Description)
	r.Uint16(&x.Flags)
	r.Uint8(&x.PermissionLevel)
	r.Int32(&aliasOffset)
	if aliasOffset >= 0 {
		x.Aliases = ctx.Enums[aliasOffset].Options
	}
	FuncIOSlice(r, &x.Overloads, ctx.CommandOverload)
}

// CommandOverload represents an overload of a command. This overload can be compared to function overloading
// in languages such as java. It represents a single usage of the command. A command may have multiple
// different overloads, which are handled differently.
type CommandOverload struct {
	// Parameters is a list of command parameters that are part of the overload. These parameters specify the
	// usage of the command when this overload is applied.
	Parameters []CommandParameter
}

// WriteCommandOverload writes a CommandOverload x to a writer.
func (ctx AvailableCommandsContext) WriteCommandOverload(r IO, x *CommandOverload) {
	FuncIOSlice(r, &x.Parameters, ctx.WriteCommandParam)
}

// CommandOverload reads a CommandOverload x from a reader.
func (ctx AvailableCommandsContext) CommandOverload(r IO, x *CommandOverload) {
	FuncIOSlice(r, &x.Parameters, ctx.CommandParam)
}

const (
	CommandArgValid    = 0x100000
	CommandArgEnum     = 0x200000
	CommandArgSuffixed = 0x1000000
	CommandArgSoftEnum = 0x4000000

	CommandArgTypeInt             = 1
	CommandArgTypeFloat           = 3
	CommandArgTypeValue           = 4
	CommandArgTypeWildcardInt     = 5
	CommandArgTypeOperator        = 6
	CommandArgTypeCompareOperator = 7
	CommandArgTypeTarget          = 8
	CommandArgTypeWildcardTarget  = 10
	CommandArgTypeFilepath        = 17
	CommandArgTypeIntegerRange    = 23
	CommandArgTypeEquipmentSlots  = 43
	CommandArgTypeString          = 44
	CommandArgTypeBlockPosition   = 52
	CommandArgTypePosition        = 53
	CommandArgTypeMessage         = 55
	CommandArgTypeRawText         = 58
	CommandArgTypeJSON            = 62
	CommandArgTypeBlockStates     = 71
	CommandArgTypeCommand         = 74
)

const (
	// ParamOptionCollapseEnum specifies if the enum (only if the Type is actually an enum type. If not,
	// setting this to true has no effect) should be collapsed. This means that the options of the enum are
	// never shown in the actual usage of the command, but only as auto-completion, like it automatically does
	// with enums that have a big amount of options. To illustrate, it can make
	// <false|true|yes|no> <$Name: bool>.
	ParamOptionCollapseEnum = iota + 1
	ParamOptionHasSemanticConstraint
	ParamOptionAsChainedCommand
)

// CommandParameter represents a single parameter of a command overload, which accepts a certain type of input
// values. It has a name and a type which show up client-side when a player is entering the command.
type CommandParameter struct {
	// Name is the name of the command parameter. It shows up in the usage like <$Name: $Type>, with the
	// exception of enum types, which show up simply as a list of options if the list is short enough and
	// Options is set to false.
	Name string
	// Type is a rather odd combination of type(flag)s that result in a certain parameter type to show up
	// client-side. It is a combination of the flags above. The basic types must be combined with the
	// ArgumentTypeFlagBasic flag (and integers with a suffix ArgumentTypeFlagSuffixed), whereas enums are
	// combined with the ArgumentTypeFlagEnum flag.
	Type uint32
	// Optional specifies if the command parameter is optional to enter. Note that no non-optional parameter
	// should ever be present in a command overload after an optional parameter. When optional, the parameter
	// shows up like so: [$Name: $Type], whereas when mandatory, it shows up like so: <$Name: $Type>.
	Optional bool
	// Options holds a combinations of options that additionally apply to the command parameter. The list of
	// options can be found above.
	Options byte

	// Enum is the enum of the parameter if it should be of the type enum. If non-empty, the parameter will
	// be treated as an enum and show up as such client-side.
	Enum CommandEnum
	// Suffix is the suffix of the parameter if it should receive one. Note that only integer argument types
	// are able to receive a suffix, and so the type, if Suffix is a non-empty string, will always be an
	// integer.
	Suffix string
}

// WriteCommandParam writes a CommandParameter x to Writer w, using the enum indices and suffix indices
// to translate the respective values to the offset in the buffer.
func (ctx AvailableCommandsContext) WriteCommandParam(w IO, x *CommandParameter) {
	if x.Enum.Dynamic {
		x.Type = CommandArgSoftEnum | CommandArgValid | uint32(ctx.DynamicEnumIndices[x.Enum.Type])
	} else if len(x.Enum.Options) != 0 {
		x.Type = CommandArgEnum | CommandArgValid | uint32(ctx.EnumIndices[x.Enum.Type])
	} else if x.Suffix != "" {
		x.Type = CommandArgSuffixed | uint32(ctx.SuffixIndices[x.Suffix])
	}
	w.String(&x.Name)
	w.Uint32(&x.Type)
	w.Bool(&x.Optional)
	w.Uint8(&x.Options)
}

// CommandParam reads a CommandParam x from Buffer src using the enums and suffixes passed to translate
// offsets to their respective values. CommandParam does not handle soft/dynamic enums. The caller is
// responsible to do this itself.
func (ctx AvailableCommandsContext) CommandParam(r IO, x *CommandParameter) {
	r.String(&x.Name)
	r.Uint32(&x.Type)
	r.Bool(&x.Optional)
	r.Uint8(&x.Options)

	// We explicitly do not do the soft enums anything here, as we haven't yet read the soft enums. The packet
	// read method will have to do this itself.
	if x.Type&CommandArgEnum != 0 {
		offset := x.Type & 0xffff
		x.Enum = ctx.Enums[offset]
	} else if x.Type&CommandArgSuffixed != 0 {
		offset := x.Type & 0xffff
		x.Suffix = ctx.Suffixes[offset]
	}
}

// CommandEnum represents an enum in a command usage. The enum typically has a type and a set of options that
// are valid. A value that is not one of the options results in a failure during execution.
type CommandEnum struct {
	// Type is the type of the command enum. The type will show up in the command usage as the type of the
	// argument if it has a certain amount of arguments, or when Options is set to true in the
	// command holding the enum.
	Type string
	// Options is a list of options that are valid for the client to submit to the command. They will be able
	// to be auto-completed and show up as options client-side.
	Options []string
	// Dynamic specifies if the command enum is considered dynamic. If set to true, it is written differently
	// and may be updated during runtime as a result using the UpdateSoftEnum packet.
	Dynamic bool
}

// Marshal encodes/decodes a CommandEnum as a "soft" (dynamic) enum.
func (x *CommandEnum) Marshal(r IO) {
	x.Dynamic = true
	r.String(&x.Type)
	FuncSlice(r, &x.Options, r.String)
}

// WriteEnum writes a CommandEnum x to a writer.
func (ctx AvailableCommandsContext) WriteEnum(w IO, x *CommandEnum) {
	w.String(&x.Type)
	FuncIOSlice(w, &x.Options, ctx.writeEnumOption)
}

// Enum reads a CommandEnum x from a reader.
func (ctx AvailableCommandsContext) Enum(r IO, x *CommandEnum) {
	r.String(&x.Type)
	FuncIOSlice(r, &x.Options, ctx.enumOption)
}

// writeEnumOption writes an enum option to w using the value indices passed. It is written as a
// byte/uint16/uint32 depending on the size of the value indices map.
func (ctx AvailableCommandsContext) writeEnumOption(w IO, option *string) {
	l := len(ctx.EnumValueIndices)
	switch {
	case l <= math.MaxUint8:
		val := byte(ctx.EnumValueIndices[*option])
		w.Uint8(&val)
	case l <= math.MaxUint16:
		val := uint16(ctx.EnumValueIndices[*option])
		w.Uint16(&val)
	default:
		val := uint32(ctx.EnumValueIndices[*option])
		w.Uint32(&val)
	}
}

// enumOption reads an enum option from buf using the enum values passed. The option is written as a
// byte/uint16/uint32, depending on the size of the enumValues slice.
func (ctx AvailableCommandsContext) enumOption(r IO, option *string) {
	l := len(ctx.EnumValues)
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
	*option = ctx.EnumValues[index]
}

const (
	CommandEnumConstraintCheatsEnabled = iota
	CommandEnumConstraintOperatorPermissions
	CommandEnumConstraintHostPermissions
	_
)

// CommandEnumConstraint is sent in the AvailableCommands packet to limit what values of an enum may be used
// taking in account things such as whether cheats are enabled.
type CommandEnumConstraint struct {
	// EnumOption is the option in an enum that the constraints should be applied to.
	EnumOption string
	// EnumName is the name of the enum of which the EnumOption above should be constrained.
	EnumName string
	// Constraints is a list of constraints that should be applied to the enum option. It is one of the values
	// found above.
	Constraints []byte
}

// EnumConstraint reads a CommandEnumConstraint x from Buffer src using the enums and enum values passed.
func (ctx AvailableCommandsContext) EnumConstraint(r IO, x *CommandEnumConstraint) {
	var enumValueIndex, enumIndex uint32
	r.Uint32(&enumValueIndex)
	r.Uint32(&enumIndex)

	x.EnumOption = ctx.EnumValues[enumValueIndex]
	x.EnumName = ctx.Enums[enumIndex].Type

	r.ByteSlice(&x.Constraints)
}

// WriteEnumConstraint writes a CommandEnumConstraint x to Writer w using the enum (value) indices passed.
func (ctx AvailableCommandsContext) WriteEnumConstraint(w IO, x *CommandEnumConstraint) {
	enumValueIndex, enumIndex := uint32(ctx.EnumValueIndices[x.EnumOption]), uint32(ctx.EnumIndices[x.EnumName])
	w.Uint32(&enumValueIndex)
	w.Uint32(&enumIndex)
	w.ByteSlice(&x.Constraints)
}

const (
	CommandOriginPlayer = iota
	CommandOriginBlock
	CommandOriginMinecartBlock
	CommandOriginDevConsole
	CommandOriginTest
	CommandOriginAutomationPlayer
	CommandOriginClientAutomation
	CommandOriginDedicatedServer
	CommandOriginEntity
	CommandOriginVirtual
	CommandOriginGameArgument
	CommandOriginEntityServer
	CommandOriginPrecompiled
	CommandOriginGameDirectorEntityServer
	CommandOriginScript
	CommandOriginExecutor
)

// CommandOrigin holds data that identifies the origin of the requesting of a command. It holds several
// fields that may be used to get specific information.
// When sent in a CommandRequest packet, the same CommandOrigin should be sent in a CommandOutput packet.
type CommandOrigin struct {
	// Origin is one of the values above that specifies the origin of the command. The origin may change,
	// depending on what part of the client actually called the command. The command may be issued by a
	// websocket server, for example.
	Origin uint32
	// UUID is the UUID of the command called. This UUID is a bit odd as it is not specified by the server. It
	// is not clear what exactly this UUID is meant to identify, but it is unique for each command called.
	UUID uuid.UUID
	// RequestID is an ID that identifies the request of the client. The server should send a CommandOrigin
	// with the same request ID to ensure it can be matched with the request by the caller of the command.
	// This is especially important for websocket servers and it seems that this field is only non-empty for
	// these websocket servers.
	RequestID string
	// PlayerUniqueID is an ID that identifies the player, the same as the one found in the AdventureSettings
	// packet. Filling it out with 0 seems to work.
	// PlayerUniqueID is only written if Origin is CommandOriginDevConsole or CommandOriginTest.
	PlayerUniqueID int64
}

// CommandOriginData reads/writes a CommandOrigin x using IO r.
func CommandOriginData(r IO, x *CommandOrigin) {
	r.Varuint32(&x.Origin)
	r.UUID(&x.UUID)
	r.String(&x.RequestID)
	if x.Origin == CommandOriginDevConsole || x.Origin == CommandOriginTest {
		r.Varint64(&x.PlayerUniqueID)
	}
}

// CommandOutputMessage represents a message sent by a command that holds the output of one of the commands
// executed.
type CommandOutputMessage struct {
	// Success indicates if the output message was one of a successful command execution. If set to true, the
	// output message is by default coloured white, whereas if set to false, the message is by default
	// coloured red.
	Success bool
	// Message is the message that is sent to the client in the chat window. It may either be simply a
	// message or a translated built-in string like 'commands.tp.success.coordinates', combined with specific
	// parameters below.
	Message string
	// Parameters is a list of parameters that serve to supply the message sent with additional information,
	// such as the position that a player was teleported to or the effect that was applied to an entity.
	// These parameters only apply for the Minecraft built-in command output.
	Parameters []string
}

// Marshal encodes/decodes a CommandOutputMessage.
func (x *CommandOutputMessage) Marshal(r IO) {
	r.Bool(&x.Success)
	r.String(&x.Message)
	FuncSlice(r, &x.Parameters, r.String)
}

// enumValues runs through all commands set to the packet and collects enum values and a map of indices
// indexed with the enum values.
func enumValues(commands []Command) (values []string, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range commands {
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
func suffixes(commands []Command) (suffixes []string, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range commands {
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
func enums(commands []Command) (enums []CommandEnum, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range commands {
		if len(command.Aliases) > 0 {
			aliasEnum := CommandEnum{Type: command.Name + "Aliases", Options: command.Aliases}
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
func dynamicEnums(commands []Command) (enums []CommandEnum, indices map[string]int) {
	indices = make(map[string]int)

	for _, command := range commands {
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
