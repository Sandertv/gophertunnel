package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ScriptCustomEvent is sent by both the client and the server. It is a way to let scripts communicate with
// the server, so that the client can let the server know it triggered an event, or the other way around.
// ScriptCustomEvent is deprecated as of 1.20.10.
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

func (pk *ScriptCustomEvent) Marshal(io protocol.IO) {
	io.String(&pk.EventName)
	io.ByteSlice(&pk.EventData)
}
