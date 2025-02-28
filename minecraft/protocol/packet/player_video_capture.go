package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerVideoCaptureActionStop = iota
	PlayerVideoCaptureActionStart
)

// PlayerVideoCapture packet is sent by the server to start or stop video recording for a player. This packet
// only works on development builds and has no effect on retail builds. When recording, the client will save
// individual frames to '/LocalCache/minecraftpe' in the format specified below.
type PlayerVideoCapture struct {
	// Action is the action to perform with the video capture. It is one of the constants above.
	Action byte
	// FrameRate is the frame rate at which the video should be recorded. It is only used when Action is
	// PlayerVideoCaptureActionStart. A higher frame rate will cause more frames to be recorded, but also
	// a noticeable increase in lag.
	FrameRate uint32
	// Unknown1, Unknown2, Unknown3 ... always 0?
	Unknown1, Unknown2, Unknown3 byte
	// FilePrefix is the prefix of the file name that will be used to save the frames. The frames will be saved
	// in the format 'FilePrefix%d.png' where %d is the frame index.
	FilePrefix string
}

// ID ...
func (*PlayerVideoCapture) ID() uint32 {
	return IDPlayerVideoCapture
}

func (pk *PlayerVideoCapture) Marshal(io protocol.IO) {
	io.Uint8(&pk.Action)
	if pk.Action == PlayerVideoCaptureActionStart {
		io.Varuint32(&pk.FrameRate)
		io.Uint8(&pk.Unknown1)
		io.Uint8(&pk.Unknown2)
		io.Uint8(&pk.Unknown3)
		io.String(&pk.FilePrefix)
	}
}
