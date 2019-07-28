package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
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
	// UnknownString ...
	UnknownString string
}

// ID ...
func (*CommandOutput) ID() uint32 {
	return IDCommandOutput
}

// Marshal ...
func (pk *CommandOutput) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteCommandOriginData(buf, pk.CommandOrigin)
	_ = binary.Write(buf, binary.LittleEndian, pk.OutputType)
	_ = protocol.WriteVaruint32(buf, pk.SuccessCount)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.OutputMessages)))
	for _, message := range pk.OutputMessages {
		_ = protocol.WriteCommandMessage(buf, message)
	}
	if pk.OutputType == 4 {
		_ = protocol.WriteString(buf, pk.UnknownString)
	}
}

// Unmarshal ...
func (pk *CommandOutput) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := chainErr(
		protocol.CommandOriginData(buf, &pk.CommandOrigin),
		binary.Read(buf, binary.LittleEndian, &pk.OutputType),
		protocol.Varuint32(buf, &pk.SuccessCount),
		protocol.Varuint32(buf, &count),
	); err != nil {
		return err
	}
	pk.OutputMessages = make([]protocol.CommandOutputMessage, count)
	for i := uint32(0); i < count; i++ {
		if err := protocol.CommandMessage(buf, &pk.OutputMessages[i]); err != nil {
			return err
		}
	}
	if pk.OutputType == 4 {
		return protocol.String(buf, &pk.UnknownString)
	}
	return nil
}
