package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ScriptCustomEvent is sent by both the client and the server. It is a way to let scripts communicate with
// the server, so that the client can let the server know it triggered an event, or the other way around.
// It is essentially an RPC kind of system.
type ScriptCustomEvent struct {
	// EventName is the name of the event. The script and the server will use this event name to identify the
	// data that is sent.
	EventName string
	// EventData is the data of the event. This data is typically a JSON encoded string, that the script is
	// able to encode and decode too.
	EventData []byte
}

// ID ...
func (*ScriptCustomEvent) ID() uint32 {
	return IDScriptCustomEvent
}

// Marshal ...
func (pk *ScriptCustomEvent) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.EventName)
	_ = protocol.WriteByteSlice(buf, pk.EventData)
}

// Unmarshal ...
func (pk *ScriptCustomEvent) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.EventName),
		protocol.ByteSlice(buf, &pk.EventData),
	)
}
