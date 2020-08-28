package protocol

import (
	"bytes"
	"github.com/go-gl/mathgl/mgl32"
)

// WriteRotation writes a rotation Vec3 to buffer dst as 3 bytes.
func WriteRotation(dst *bytes.Buffer, x mgl32.Vec3) error {
	return chainErr(
		dst.WriteByte(byte(x[0]/(360.0/256.0))),
		dst.WriteByte(byte(x[1]/(360.0/256.0))),
		dst.WriteByte(byte(x[2]/(360.0/256.0))),
	)
}
