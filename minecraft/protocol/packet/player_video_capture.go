package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerVideoCaptureActionStart = iota
	PlayerVideoCaptureActionStop
)

// PlayerVideoCapture packet is sent by the server to start or stop video recording for a player. This packet
// only works on development builds and has no effect on retail builds. When recording, the client will save
// individual frames to '/LocalCache/minecraftpe' in the format specified below.
type PlayerVideoCapture struct {
	// StartVideoCapture is the action to perform with the video capture.
	StartVideoCapture PlayerVideoCaptureAction
	// StopVideoCapture is a duplicate of StartVideoCapture.
	StopVideoCapture PlayerVideoCaptureAction
}

// ID ...
func (*PlayerVideoCapture) ID() uint32 {
	return IDPlayerVideoCapture
}

func (pk *PlayerVideoCapture) Marshal(io protocol.IO) {
	protocol.Single(io, &pk.StartVideoCapture)
	protocol.Single(io, &pk.StopVideoCapture)
}

// PlayerVideoCaptureAction is an action to perform with the video capture of a player.
type PlayerVideoCaptureAction struct {
	// Action is the action to perform with the video capture. It is one of the constants above.
	Action uint32
	// FrameRate is the frame rate at which the video should be recorded. It is only used when Action is
	// PlayerVideoCaptureActionStart. A higher frame rate will cause more frames to be recorded, but also
	// a noticeable increase in lag.
	FrameRate int32
	// FilePrefix is the prefix of the file name that will be used to save the frames. The frames will be saved
	// in the format 'FilePrefix%d.png' where %d is the frame index.
	FilePrefix string
}

func (x *PlayerVideoCaptureAction) Marshal(io protocol.IO) {
	io.Varuint32(&x.Action)
	if x.Action == PlayerVideoCaptureActionStart {
		io.Int32(&x.FrameRate)
		io.String(&x.FilePrefix)
	}
}
