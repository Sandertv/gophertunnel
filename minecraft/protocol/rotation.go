package protocol

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

// Rotation reads a rotation object from buffer src and stores it in Vec3 x. The rotation object exists out
// of 3 bytes.
func Rotation(src *bytes.Buffer, x *mgl32.Vec3) error {
	data := src.Next(3)
	if len(data) != 3 {
		return fmt.Errorf("%v: expected exactly 3 bytes for byte rotation", callFrame())
	}
	(*x)[0] = float32(data[0]) * (360.0 / 256.0)
	(*x)[1] = float32(data[1]) * (360.0 / 256.0)
	(*x)[2] = float32(data[2]) * (360.0 / 256.0)
	return nil
}

// WriteRotation writes a rotation Vec3 to buffer src as 3 bytes.
func WriteRotation(src *bytes.Buffer, x mgl32.Vec3) error {
	return chainErr(
		src.WriteByte(byte(x[0]/(360.0/256.0))),
		src.WriteByte(byte(x[1]/(360.0/256.0))),
		src.WriteByte(byte(x[2]/(360.0/256.0))),
	)
}
