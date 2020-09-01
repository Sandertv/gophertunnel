package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	CommandOutputTypeNone = iota
	CommandOutputTypeLastOutput
	CommandOutputTypeSilent
	CommandOutputTypeAllOutput
	CommandOutputTypeDataSet
)

// CommandOutput is sent by the server to the client to send text as output of a command. Most servers do not
// use this packet and instead simply send Text packets, but there is reason to send it.
// If the origin of a CommandRequest packet is not the player itself, but, for example, a websocket server,
// sending a Text packet will not do what is expected: The message should go to the websocket server, not to
// the client's chat. The CommandOutput packet will make sure the messages are relayed to the correct origin
// of the command request.
type CommandOutput struct {
	// CommandOrigin is the data specifying the origin of the command. In other words, the source that the
	// command request was from, such as the player itself or a websocket server. The client forwards the
	// messages in this packet to the right origin, depending on what is sent here.
	CommandOrigin protocol.CommandOrigin
	// OutputType specifies the type of output that is sent. The OutputType sent by vanilla games appears to
	// be 3, which seems to work.
	OutputType byte
	// SuccessCount is the amount of times that a command was executed successfully as a result of the command
	// that was requested. For servers, this is usually a rather meaningless fields, but for vanilla, this is
	// applicable for commands created with Functions.
	SuccessCount uint32
	// OutputMessages is a list of all output messages that should be sent to the player. Whether they are
	// shown or not, depends on the type of the messages.
	OutputMessages []protocol.CommandOutputMessage
	// DataSet ... TODO: Find out what this is for.
	DataSet string
}

// ID ...
func (*CommandOutput) ID() uint32 {
	return IDCommandOutput
}

// Marshal ...
func (pk *CommandOutput) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.OutputMessages))

	protocol.CommandOriginData(w, &pk.CommandOrigin)
	w.Uint8(&pk.OutputType)
	w.Varuint32(&pk.SuccessCount)
	w.Varuint32(&l)
	for _, message := range pk.OutputMessages {
		protocol.WriteCommandMessage(w, &message)
	}
	if pk.OutputType == CommandOutputTypeDataSet {
		w.String(&pk.DataSet)
	}
}

// Unmarshal ...
func (pk *CommandOutput) Unmarshal(r *protocol.Reader) {
	var count uint32
	protocol.CommandOriginData(r, &pk.CommandOrigin)
	r.Uint8(&pk.OutputType)
	r.Varuint32(&pk.SuccessCount)
	r.Varuint32(&count)
	pk.OutputMessages = make([]protocol.CommandOutputMessage, count)
	for i := uint32(0); i < count; i++ {
		protocol.CommandMessage(r, &pk.OutputMessages[i])
	}
	if pk.OutputType == CommandOutputTypeDataSet {
		r.String(&pk.DataSet)
	}
}
