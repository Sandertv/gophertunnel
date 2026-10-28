package packet

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type marshaler interface {
	Marshal(io protocol.IO)
}

// TestWireShapes pins layouts where optional fields and union selectors are easy to get wrong.
func TestWireShapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		value  marshaler
		decode func() marshaler
		prefix string
	}{
		{
			name:   "structure name without filtered name",
			value:  &StructureBlockUpdate{StructureName: "a"},
			decode: func() marshaler { return &StructureBlockUpdate{} },
			prefix: "000000" + "0161" + "00",
		},
		{
			name:   "structure name with filtered name",
			value:  &StructureBlockUpdate{StructureName: "a", FilteredStructureName: protocol.Option("b")},
			decode: func() marshaler { return &StructureBlockUpdate{} },
			prefix: "000000" + "0161" + "010162",
		},
		{
			name:   "video capture start is selector zero",
			value:  &PlayerVideoCapture{Action: PlayerVideoCaptureActionStart, FrameRate: 30, FilePrefix: "f"},
			decode: func() marshaler { return &PlayerVideoCapture{} },
			prefix: "00" + "1e000000" + "0166",
		},
		{
			name:   "video capture stop is selector one",
			value:  &PlayerVideoCapture{Action: PlayerVideoCaptureActionStop},
			decode: func() marshaler { return &PlayerVideoCapture{} },
			prefix: "01",
		},
		{
			name:   "video capture selector is a varuint32",
			value:  &PlayerVideoCapture{Action: 128},
			decode: func() marshaler { return &PlayerVideoCapture{} },
			prefix: "8001",
		},
		{
			name: "spline instruction fields are not optional",
			value: &protocol.CameraSplineInstruction{
				SplineType:       1,
				SplineIdentifier: "s",
				LoadFromJson:     true,
				ProgressKeyFrames: []protocol.CameraProgressOption{
					{EaseType: protocol.EasingTypeLinear},
				},
				Curve:           []mgl32.Vec3{},
				RotationOptions: []protocol.CameraRotationOption{},
			},
			decode: func() marshaler { return &protocol.CameraSplineInstruction{} },
			prefix: "00000000" + "01" + "00" + "01" + "00000000" + "00000000" + "066c696e656172" + "00" + "0173" + "01",
		},
		{
			name: "spline definition key frame easing is optional",
			value: &CameraSpline{Splines: []protocol.CameraSplineDefinition{{
				Name:          "a:b",
				SplineType:    protocol.SplineTypeLinear,
				ControlPoints: []mgl32.Vec3{},
				ProgressKeyFrames: []protocol.CameraSplineProgressKeyFrame{
					{},
					{EaseType: protocol.Option(int32(protocol.EasingTypeLinear))},
				},
				RotationKeyFrames: []protocol.CameraSplineRotationKeyFrame{},
			}}},
			decode: func() marshaler { return &CameraSpline{} },
			prefix: "01" + "03613a62" + "00000000" + "066c696e656172" + "00" +
				"02" + "00000000" + "00000000" + "00" + "00000000" + "00000000" + "01066c696e656172" + "00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var encoded bytes.Buffer
			tt.value.Marshal(protocol.NewWriter(&encoded, 0))
			if got := hex.EncodeToString(encoded.Bytes()); !strings.HasPrefix(got, tt.prefix) {
				t.Fatalf("encoded %s, want prefix %s", got, tt.prefix)
			}
			decoded := tt.decode()
			decoded.Marshal(protocol.NewReader(bytes.NewBuffer(encoded.Bytes()), 0, false))
			if !reflect.DeepEqual(decoded, tt.value) {
				t.Fatalf("decoded %#v, want %#v", decoded, tt.value)
			}
		})
	}
}
