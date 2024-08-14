package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerBoundDiagnostics is sent by the client to tell the server about the performance diagnostics
// of the client. It is sent by the client roughly every 500ms or 10 in-game ticks.
type ServerBoundDiagnostics struct {
	// AverageFramesPerSecond is the average amount of frames per second that the client has been
	// running at.
	AverageFramesPerSecond float32
	// AverageServerSimTickTime is the average time that the server spends simulating a single tick
	// in milliseconds.
	AverageServerSimTickTime float32
	// AverageClientSimTickTime is the average time that the client spends simulating a single tick
	// in milliseconds.
	AverageClientSimTickTime float32
	// AverageBeginFrameTime is the average time that the client spends beginning a frame in
	// milliseconds.
	AverageBeginFrameTime float32
	// AverageInputTime is the average time that the client spends processing input in milliseconds.
	AverageInputTime float32
	// AverageRenderTime is the average time that the client spends rendering in milliseconds.
	AverageRenderTime float32
	// AverageEndFrameTime is the average time that the client spends ending a frame in milliseconds.
	AverageEndFrameTime float32
	// AverageRemainderTimePercent is the average percentage of time that the client spends on
	// tasks that are not accounted for.
	AverageRemainderTimePercent float32
	// AverageUnaccountedTimePercent is the average percentage of time that the client spends on
	// unaccounted tasks.
	AverageUnaccountedTimePercent float32
}

// ID ...
func (*ServerBoundDiagnostics) ID() uint32 {
	return IDServerBoundDiagnostics
}

func (pk *ServerBoundDiagnostics) Marshal(io protocol.IO) {
	io.Float32(&pk.AverageFramesPerSecond)
	io.Float32(&pk.AverageServerSimTickTime)
	io.Float32(&pk.AverageClientSimTickTime)
	io.Float32(&pk.AverageBeginFrameTime)
	io.Float32(&pk.AverageInputTime)
	io.Float32(&pk.AverageRenderTime)
	io.Float32(&pk.AverageEndFrameTime)
	io.Float32(&pk.AverageRemainderTimePercent)
	io.Float32(&pk.AverageUnaccountedTimePercent)
}
