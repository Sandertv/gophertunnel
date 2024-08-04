package protocol

import (
	"github.com/google/uuid"
	"math"
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
	Flags uint16
	// PermissionLevel is the command permission level that the player required to execute this command. The
	// field no longer seems to serve a purpose, as the client does not handle the execution of commands
	// anymore: The permissions should be checked server-side.
	PermissionLevel byte
	// AliasesOffset is the offset to a CommandEnum that holds the values that
	// should be used as aliases for this command.
	AliasesOffset uint32
	// ChainedSubcommandOffsets is a slice of offsets that all point to a different ChainedSubcommand from the
	// ChainedSubcommands slice in the AvailableCommands packet.
	ChainedSubcommandOffsets []uint16
	// Overloads is a list of command overloads that specify the ways in which a command may be executed. The
	// overloads may be completely different.
	Overloads []CommandOverload
}

func (c *Command) Marshal(r IO) {
	r.String(&c.Name)
	r.String(&c.Description)
	r.Uint16(&c.Flags)
	r.Uint8(&c.PermissionLevel)
	r.Uint32(&c.AliasesOffset)
	FuncSlice(r, &c.ChainedSubcommandOffsets, r.Uint16)
	Slice(r, &c.Overloads)
}

// CommandOverload represents an overload of a command. This overload can be compared to function overloading
// in languages such as java. It represents a single usage of the command. A command may have multiple
// different overloads, which are handled differently.
type CommandOverload struct {
	// Chaining determines if the parameters use chained subcommands or not.
	Chaining bool
	// Parameters is a list of command parameters that are part of the overload. These parameters specify the
	// usage of the command when this overload is applied.
	Parameters []CommandParameter
}

func (c *CommandOverload) Marshal(r IO) {
	r.Bool(&c.Chaining)
	Slice(r, &c.Parameters)
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
	CommandArgTypeEquipmentSlots  = 47
	CommandArgTypeString          = 56
	CommandArgTypeBlockPosition   = 64
	CommandArgTypePosition        = 65
	CommandArgTypeMessage         = 67
	CommandArgTypeRawText         = 70
	CommandArgTypeJSON            = 74
	CommandArgTypeBlockStates     = 83
	CommandArgTypeCommand         = 86
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
}

func (c *CommandParameter) Marshal(r IO) {
	r.String(&c.Name)
	r.Uint32(&c.Type)
	r.Bool(&c.Optional)
	r.Uint8(&c.Options)
}

// CommandEnum represents an enum in a command usage. The enum typically has a type and a set of options that
// are valid. A value that is not one of the options results in a failure during execution.
type CommandEnum struct {
	// Type is the type of the command enum. The type will show up in the command usage as the type of the
	// argument if it has a certain amount of arguments, or when Options is set to true in the
	// command holding the enum.
	Type string
	// ValueIndices holds a list of indices that point to the EnumValues slice in the
	// AvailableCommandsPacket. These represent the options of the enum.
	ValueIndices []uint
}

// CommandEnumContext holds context required for encoding command enums.
type CommandEnumContext struct {
	EnumValues []string
}

// Marshal encodes/decodes a CommandEnum.
func (ctx CommandEnumContext) Marshal(r IO, x *CommandEnum) {
	r.String(&x.Type)
	FuncIOSlice(r, &x.ValueIndices, ctx.enumOption)
}

// enumOption writes/reads a command enum option as a byte/uint16/uint32,
// depending on the amount of enum values.
func (ctx CommandEnumContext) enumOption(r IO, opt *uint) {
	n := len(ctx.EnumValues)
	switch {
	case n <= math.MaxUint8:
		val := byte(*opt)
		r.Uint8(&val)
		*opt = uint(val)
	case n <= math.MaxUint16:
		val := uint16(*opt)
		r.Uint16(&val)
		*opt = uint(val)
	default:
		val := uint32(*opt)
		r.Uint32(&val)
		*opt = uint(val)
	}
}

// ChainedSubcommand represents a subcommand that can have chained commands, such as /execute which allows you to run
// another command as another entity or at a different position etc.
type ChainedSubcommand struct {
	// Name is the name of the chained subcommand and shows up in the list as a regular subcommand enum.
	Name string
	// Values contains the index and parameter type of the chained subcommand.
	Values []ChainedSubcommandValue
}

func (x *ChainedSubcommand) Marshal(r IO) {
	r.String(&x.Name)
	Slice(r, &x.Values)
}

// ChainedSubcommandValue represents the value for a chained subcommand argument.
type ChainedSubcommandValue struct {
	// Index is the index of the argument in the ChainedSubcommandValues slice from the AvailableCommands packet. This is
	// then used to set the type specified by the Value field below.
	Index uint16
	// Value is a combination of the flags above and specified the type of argument. Unlike regular parameter types,
	// this should NOT contain any of the special flags (valid, enum, suffixed or soft enum) but only the basic types.
	Value uint16
}

func (x *ChainedSubcommandValue) Marshal(r IO) {
	r.Uint16(&x.Index)
	r.Uint16(&x.Value)
}

// DynamicEnum is an enum variant that can have its options changed during runtime,
// without sending a new AvailableCommands packet.
type DynamicEnum struct {
	// Type is the type of the command enum. The type will show up in the command usage as the type of the
	// argument if it has a certain amount of arguments, or when Options is set to true in the
	// command holding the enum.
	Type string
	// Values is a slice of possible options for the enum.
	Values []string
}

// Marshal encodes/decodes a DynamicEnum.
func (x *DynamicEnum) Marshal(r IO) {
	r.String(&x.Type)
	FuncSlice(r, &x.Values, r.String)
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
	// EnumValueIndex points to an enum value in the AvailableCommands packet that this
	// constraint should apply to.
	EnumValueIndex uint32
	// EnumIndex points to an enum in the AvailableCommands packet to which this
	// constraint should apply to.
	EnumIndex uint32
	// Constraints holds a slice of constraints as present in the constants above.
	Constraints []byte
}

// Marshal encodes/decodes a CommandEnumConstraint.
func (c *CommandEnumConstraint) Marshal(r IO) {
	r.Uint32(&c.EnumValueIndex)
	r.Uint32(&c.EnumIndex)
	r.ByteSlice(&c.Constraints)
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
	// UUID is a unique identifier for every instantiation of a command.
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
