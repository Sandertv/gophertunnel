package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
)

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
	Flags byte
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

// CommandOverload represents an overload of a command. This overload can be compared to function overloading
// in languages such as java. It represents a single usage of the command. A command may have multiple
// different overloads, which are handled differently.
type CommandOverload struct {
	// Parameters is a list of command parameters that are part of the overload. These parameters specify the
	// usage of the command when this overload is applied.
	Parameters []CommandParameter
}

const (
	CommandArgValid    = 0x100000
	CommandArgEnum     = 0x200000
	CommandArgSuffixed = 0x1000000
	CommandArgSoftEnum = 0x4000000

	CommandArgTypeInt         = 1
	CommandArgTypeFloat       = 2
	CommandArgTypeValue       = 3
	CommandArgTypeWildcardInt = 4
	CommandArgTypeOperator    = 5
	CommandArgTypeTarget      = 6

	CommandArgTypeFilepath = 0x0e
	CommandArgTypeString   = 0x1b
	CommandArgTypePosition = 0x1d
	CommandArgTypeMessage  = 0x20
	CommandArgTypeRawText  = 0x22
	CommandArgTypeJSON     = 0x25
	CommandArgTypeCommand  = 0x2c
)
const (
	// ParamOptionCollapseEnum specifies if the enum (only if the Type is actually an enum type. If not,
	// setting this to true has no effect) should be collapsed. This means that the options of the enum are
	// never shown in the actual usage of the command, but only as auto-completion, like it automatically does
	// with enums that have a big amount of options. To illustrate, it can make
	// <false|true|yes|no> <$Name: bool>.
	ParamOptionCollapseEnum = iota + 1
	ParamOptionHasSemanticConstraint
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

// CommandMessage reads a CommandOutputMessage x from Buffer src.
func CommandMessage(src *bytes.Buffer, x *CommandOutputMessage) error {
	var count uint32
	if err := chainErr(
		binary.Read(src, binary.LittleEndian, &x.Success),
		String(src, &x.Message),
		Varuint32(src, &count),
	); err != nil {
		return err
	}
	x.Parameters = make([]string, count)
	for i := uint32(0); i < count; i++ {
		if err := String(src, &x.Parameters[i]); err != nil {
			return err
		}
	}
	return nil
}

// WriteCommandMessage writes a CommandOutputMessage x to Buffer dst.
func WriteCommandMessage(dst *bytes.Buffer, x CommandOutputMessage) error {
	if err := chainErr(
		binary.Write(dst, binary.LittleEndian, x.Success),
		WriteString(dst, x.Message),
		WriteVaruint32(dst, uint32(len(x.Parameters))),
	); err != nil {
		return err
	}
	for _, param := range x.Parameters {
		if err := WriteString(dst, param); err != nil {
			return err
		}
	}
	return nil
}

// WriteCommandOriginData writes a CommandOrigin x to Buffer dst.
func WriteCommandOriginData(dst *bytes.Buffer, x CommandOrigin) error {
	if err := chainErr(
		WriteVaruint32(dst, x.Origin),
		WriteUUID(dst, x.UUID),
		WriteString(dst, x.RequestID),
	); err != nil {
		return err
	}
	if x.Origin == CommandOriginDevConsole || x.Origin == CommandOriginTest {
		return WriteVarint64(dst, x.PlayerUniqueID)
	}
	return nil
}

// CommandOriginData reads a CommandOrigin x from Buffer src.
func CommandOriginData(src *bytes.Buffer, x *CommandOrigin) error {
	if err := chainErr(
		Varuint32(src, &x.Origin),
		UUID(src, &x.UUID),
		String(src, &x.RequestID),
	); err != nil {
		return err
	}
	if x.Origin == CommandOriginDevConsole || x.Origin == CommandOriginTest {
		return Varint64(src, &x.PlayerUniqueID)
	}
	return nil
}

// WriteCommandData writes a Command x to Buffer dst, using the enum indices and suffix indices passed to
// translate enums and suffixes to the indices that they're written in in the buffer.
func WriteCommandData(dst *bytes.Buffer, x Command, enumIndices map[string]int, suffixIndices map[string]int, dynamicEnumIndices map[string]int) error {
	alias := int32(-1)
	if len(x.Aliases) != 0 {
		alias = int32(enumIndices[x.Name+"Aliases"])
	}
	if err := chainErr(
		WriteString(dst, x.Name),
		WriteString(dst, x.Description),
		binary.Write(dst, binary.LittleEndian, x.Flags),
		binary.Write(dst, binary.LittleEndian, x.PermissionLevel),
		binary.Write(dst, binary.LittleEndian, alias),
	); err != nil {
		return err
	}
	if err := WriteVaruint32(dst, uint32(len(x.Overloads))); err != nil {
		return err
	}
	for _, overload := range x.Overloads {
		if err := WriteVaruint32(dst, uint32(len(overload.Parameters))); err != nil {
			return err
		}
		for _, param := range overload.Parameters {
			if err := WriteCommandParam(dst, param, enumIndices, suffixIndices, dynamicEnumIndices); err != nil {
				return err
			}
		}
	}
	return nil
}

// CommandData reads a Command x from Buffer src using the enums and suffixes passed to match indices with
// the values these slices hold.
func CommandData(src *bytes.Buffer, x *Command, enums []CommandEnum, suffixes []string) error {
	var (
		overloadCount, paramCount uint32
		aliasOffset               int32
	)
	if err := chainErr(
		String(src, &x.Name),
		String(src, &x.Description),
		binary.Read(src, binary.LittleEndian, &x.Flags),
		binary.Read(src, binary.LittleEndian, &x.PermissionLevel),
		binary.Read(src, binary.LittleEndian, &aliasOffset),
	); err != nil {
		return err
	}
	if aliasOffset >= 0 {
		if len(enums) <= int(aliasOffset) {
			return fmt.Errorf("invalid alias offset %v, expected lower than or equal to %v", aliasOffset, len(enums))
		}
		x.Aliases = enums[aliasOffset].Options
	}
	if err := Varuint32(src, &overloadCount); err != nil {
		return err
	}
	x.Overloads = make([]CommandOverload, overloadCount)
	for i := uint32(0); i < overloadCount; i++ {
		if err := Varuint32(src, &paramCount); err != nil {
			return err
		}
		x.Overloads[i].Parameters = make([]CommandParameter, paramCount)
		for j := uint32(0); j < paramCount; j++ {
			if err := CommandParam(src, &x.Overloads[i].Parameters[j], enums, suffixes); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteCommandParam writes a CommandParameter x to Buffer dst, using the enum indices and suffix indices
// to translate the respective values to the offset in the buffer.
func WriteCommandParam(dst *bytes.Buffer, x CommandParameter, enumIndices map[string]int, suffixIndices map[string]int, dynamicEnumIndices map[string]int) error {
	if x.Enum.Dynamic {
		x.Type = CommandArgSoftEnum | CommandArgValid | uint32(dynamicEnumIndices[x.Enum.Type])
	} else if len(x.Enum.Options) != 0 {
		x.Type = CommandArgEnum | CommandArgValid | uint32(enumIndices[x.Enum.Type])
	} else if x.Suffix != "" {
		x.Type = CommandArgSuffixed | uint32(suffixIndices[x.Suffix])
	}
	return chainErr(
		WriteString(dst, x.Name),
		binary.Write(dst, binary.LittleEndian, x.Type),
		binary.Write(dst, binary.LittleEndian, x.Optional),
		binary.Write(dst, binary.LittleEndian, x.Options),
	)
}

// CommandParam reads a CommandParam x from Buffer src using the enums and suffixes passed to translate
// offsets to their respective values. CommandParam does not handle soft/dynamic enums. The caller is
// responsible to do this itself.
func CommandParam(src *bytes.Buffer, x *CommandParameter, enums []CommandEnum, suffixes []string) error {
	if err := chainErr(
		String(src, &x.Name),
		binary.Read(src, binary.LittleEndian, &x.Type),
		binary.Read(src, binary.LittleEndian, &x.Optional),
		binary.Read(src, binary.LittleEndian, &x.Options),
	); err != nil {
		return err
	}
	if x.Type&CommandArgSoftEnum != 0 {
		// We explicitly do not do anything here, as we haven't yet read the soft enums. The packet read
		// method will have to do this itself.
	} else if x.Type&CommandArgEnum != 0 {
		offset := int(x.Type & 0xffff)
		if len(enums) <= offset {
			return fmt.Errorf("invalid enum offset %v, expected one lower than or equal to %v", offset, len(enums))
		}
		x.Enum = enums[offset]
	} else if x.Type&CommandArgSuffixed != 0 {
		offset := int(x.Type & 0xffff)
		if len(suffixes) <= offset {
			return fmt.Errorf("invalid suffix offset %v, expected one lower than or equal to %v", offset, len(suffixes))
		}
		x.Suffix = suffixes[offset]
	}
	return nil
}
