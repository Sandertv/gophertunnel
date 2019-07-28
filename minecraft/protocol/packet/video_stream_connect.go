package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	VideoStreamActionStart = iota
	VideoStreamActionStop
)

// VideoStreamConnect is sent by the server to make the client start video streaming. Essentially, what this
// means is that the client will start sending screenshots at a rate specified to the websocket server it
// connects to with this packet.
// In version 1.11, this packet is blocked: The client sends a client-side message that video streaming is
// not supported on the device.
type VideoStreamConnect struct {
	// ServerURI is the URI to make the client connect to. It can be, for example, 'localhost:8000/ws' to
	// connect to a websocket server on the localhost at port 8000.
	// Screenshots will be sent to this websocket server.
	ServerURI string
	// FrameSendFrequency is the frequency at which the client should send frames, AKA screenshots to the
	// websocket server. The exact unit of this value is not clear: The command in-game does not work, so
	// it is impossible to find out.
	FrameSendFrequency float32
	// ActionType is the type of the action to execute. It is either VideoStreamActionStart or
	// VideoStreamActionStop to start or stop the video streaming respectively.
	ActionType byte
	// ResolutionX is the width in pixels of the 'screenshots' sent to the websocket. This is the resolution
	// of the video on the X axis.
	ResolutionX int32
	// ResolutionY is the height in pixels of the 'screenshots' sent to the websocket. This is the resolution
	// of the video on the Y axis.
	ResolutionY int32
}

// ID ...
func (*VideoStreamConnect) ID() uint32 {
	return IDVideoStreamConnect
}

// Marshal ...
func (pk *VideoStreamConnect) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteString(buf, pk.ServerURI)
	_ = protocol.WriteFloat32(buf, pk.FrameSendFrequency)
	_ = binary.Write(buf, binary.LittleEndian, pk.ActionType)
	_ = binary.Write(buf, binary.LittleEndian, pk.ResolutionX)
	_ = binary.Write(buf, binary.LittleEndian, pk.ResolutionY)
}

// Unmarshal ...
func (pk *VideoStreamConnect) Unmarshal(buf *bytes.Buffer) error {
	return chainErr(
		protocol.String(buf, &pk.ServerURI),
		protocol.Float32(buf, &pk.FrameSendFrequency),
		binary.Read(buf, binary.LittleEndian, &pk.ActionType),
		binary.Read(buf, binary.LittleEndian, &pk.ResolutionX),
		binary.Read(buf, binary.LittleEndian, &pk.ResolutionY),
	)
}
