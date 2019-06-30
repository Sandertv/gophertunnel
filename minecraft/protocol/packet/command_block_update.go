package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CommandBlockImpulse = iota
	CommandBlockRepeat
	CommandBlockChain
)

// CommandBlockUpdate is sent by the server to update a command block at a specific position. The command
// block may be either a physical block or an entity.
type CommandBlockUpdate struct {
	// Block specifies if the command block updated was an actual physical block. If false, the command block
	// is in a minecart and has an entity runtime ID instead.
	Block bool

	// Position is the position of the command block updated. It is only set if Block is set to true. Nothing
	// happens if no command block is set at this position.
	Position protocol.BlockPos
	// Mode is the mode of the command block. It is either CommandBlockImpulse, CommandBlockChain or
	// CommandBlockRepeat. It is only set if Block is set to true.
	Mode uint32
	// NeedsRedstone specifies if the command block needs to be powered by redstone to be activated. If false,
	// the command block is always active. The field is only set if Block is set to true.
	NeedsRedstone bool
	// Conditional specifies the behaviour of the command block if the command block before it (the opposite
	// side of the direction the arrow if facing) fails to execute. If set to false, it will activate at all
	// times, whereas if set to true, it will activate only if the previous command block executed
	// successfully. The field is only set if Block is set to true.
	Conditional bool

	// MinecartEntityRuntimeID is the runtime ID of the minecart entity carrying the command block that is
	// updated. It is set only if Block is set to false.
	MinecartEntityRuntimeID uint64

	// Command is the command currently entered in the command block. This is the command that is executed
	// when the command block is activated.
	Command string
	// LastOutput is the output of the last command executed by the command block. It may be left empty to
	// show simply no output at all, in combination with setting ShouldTrackOutput to false.
	LastOutput string
	// Name is the name of the command block updated. If not empty, it will show this name hovering above the
	// command block when hovering over the block with the cursor.
	Name string
	// ShouldTrackOutput specifies if the command block tracks output. If set to false, the output box won't
	// be shown within the command block.
	ShouldTrackOutput bool
}

// ID ...
func (*CommandBlockUpdate) ID() uint32 {
	return IDCommandBlockUpdate
}

// Marshal ...
func (pk *CommandBlockUpdate) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Block)
	if pk.Block {
		_ = protocol.WriteUBlockPosition(buf, pk.Position)
		_ = protocol.WriteVaruint32(buf, pk.Mode)
		_ = binary.Write(buf, binary.LittleEndian, pk.NeedsRedstone)
		_ = binary.Write(buf, binary.LittleEndian, pk.Conditional)
	} else {
		_ = protocol.WriteVaruint64(buf, pk.MinecartEntityRuntimeID)
	}
	_ = protocol.WriteString(buf, pk.Command)
	_ = protocol.WriteString(buf, pk.LastOutput)
	_ = protocol.WriteString(buf, pk.Name)
	_ = binary.Write(buf, binary.LittleEndian, pk.ShouldTrackOutput)
}

// Unmarshal ...
func (pk *CommandBlockUpdate) Unmarshal(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.LittleEndian, &pk.Block); err != nil {
		return err
	}
	if pk.Block {
		if err := chainErr(
			protocol.UBlockPosition(buf, &pk.Position),
			protocol.Varuint32(buf, &pk.Mode),
			binary.Read(buf, binary.LittleEndian, &pk.NeedsRedstone),
			binary.Read(buf, binary.LittleEndian, &pk.Conditional),
		); err != nil {
			return err
		}
	} else {
		if err := protocol.Varuint64(buf, &pk.MinecartEntityRuntimeID); err != nil {
			return err
		}
	}
	return chainErr(
		protocol.String(buf, &pk.Command),
		protocol.String(buf, &pk.LastOutput),
		protocol.String(buf, &pk.Name),
		binary.Read(buf, binary.LittleEndian, &pk.ShouldTrackOutput),
	)
}
