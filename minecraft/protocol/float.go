package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

// WriteFloat32 writes a float32 to Buffer dst, by first converting it to a uint32.
func WriteFloat32(dst *bytes.Buffer, x float32) error {
	if err := binary.Write(dst, binary.LittleEndian, math.Float32bits(x)); err != nil {
		return wrap(err)
	}
	return nil
}

// WriteVec3 writes an mgl32.Vec3 (float32 vector) to Buffer dst, writing each of the float32s separately.
func WriteVec3(dst *bytes.Buffer, x mgl32.Vec3) error {
	return chainErr(
		WriteFloat32(dst, x[0]),
		WriteFloat32(dst, x[1]),
		WriteFloat32(dst, x[2]),
	)
}

// WriteVec2 writes an mgl32.Vec2 (float32 vector) to Buffer dst, writing each of the float32s separately.
func WriteVec2(dst *bytes.Buffer, x mgl32.Vec2) error {
	return chainErr(
		WriteFloat32(dst, x[0]),
		WriteFloat32(dst, x[1]),
	)
}
