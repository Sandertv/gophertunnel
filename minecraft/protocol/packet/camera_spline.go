package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CameraSpline is sent by the server to define camera spline paths.
type CameraSpline struct {
	// Splines is a list of camera spline definitions.
	Splines []protocol.CameraSplineDefinition
}

// ID ...
func (*CameraSpline) ID() uint32 {
	return IDCameraSpline
}

func (pk *CameraSpline) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Splines)
}
