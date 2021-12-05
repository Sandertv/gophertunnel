package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CommandBlockImpulse = iota
	CommandBlockRepeating
	CommandBlockChain
)

// CommandBlockUpdate is sent by the client to update a command block at a specific position. The command
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
	// TickDelay is the delay in ticks between executions of a command block, if it is a repeating command
	// block.
	TickDelay int32
	// ExecuteOnFirstTick specifies if the command block should execute on the first tick, AKA as soon as the
	// command block is enabled.
	ExecuteOnFirstTick bool
}

// ID ...
func (*CommandBlockUpdate) ID() uint32 {
	return IDCommandBlockUpdate
}

// Marshal ...
func (pk *CommandBlockUpdate) Marshal(w *protocol.Writer) {
	w.Bool(&pk.Block)
	if pk.Block {
		w.UBlockPos(&pk.Position)
		w.Varuint32(&pk.Mode)
		w.Bool(&pk.NeedsRedstone)
		w.Bool(&pk.Conditional)
	} else {
		w.Varuint64(&pk.MinecartEntityRuntimeID)
	}
	w.String(&pk.Command)
	w.String(&pk.LastOutput)
	w.String(&pk.Name)
	w.Bool(&pk.ShouldTrackOutput)
	w.Int32(&pk.TickDelay)
	w.Bool(&pk.ExecuteOnFirstTick)
}

// Unmarshal ...
func (pk *CommandBlockUpdate) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.Block)
	if pk.Block {
		r.UBlockPos(&pk.Position)
		r.Varuint32(&pk.Mode)
		r.Bool(&pk.NeedsRedstone)
		r.Bool(&pk.Conditional)
	} else {
		r.Varuint64(&pk.MinecartEntityRuntimeID)
	}
	r.String(&pk.Command)
	r.String(&pk.LastOutput)
	r.String(&pk.Name)
	r.Bool(&pk.ShouldTrackOutput)
	r.Int32(&pk.TickDelay)
	r.Bool(&pk.ExecuteOnFirstTick)
}
