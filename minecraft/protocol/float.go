package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

// Float32 reads a float32 from Buffer src, setting the result to the pointer to a float32 passed.
func Float32(src *bytes.Buffer, x *float32) error {
	var bits uint32
	if err := binary.Read(src, binary.LittleEndian, &bits); err != nil {
		return wrap(err)
	}
	*x = math.Float32frombits(bits)
	return nil
}

// WriteFloat32 writes a float32 to Buffer dst, by first converting it to a uint32.
func WriteFloat32(dst *bytes.Buffer, x float32) error {
	if err := binary.Write(dst, binary.LittleEndian, math.Float32bits(x)); err != nil {
		return wrap(err)
	}
	return nil
}

// Vec3 reads an mgl32.Vec3 (float32 vector) from Buffer src, setting the result to the pointer to an
// mgl32.Vec3 passed.
func Vec3(src *bytes.Buffer, x *mgl32.Vec3) error {
	return chainErr(
		Float32(src, &(*x)[0]),
		Float32(src, &(*x)[1]),
		Float32(src, &(*x)[2]),
	)
}

// WriteVec3 writes an mgl32.Vec3 (float32 vector) to Buffer dst, writing each of the float32s separately.
func WriteVec3(dst *bytes.Buffer, x mgl32.Vec3) error {
	return chainErr(
		WriteFloat32(dst, x[0]),
		WriteFloat32(dst, x[1]),
		WriteFloat32(dst, x[2]),
	)
}

// Vec2 reads an mgl32.Vec2 (float32 vector) from Buffer src, setting the result to the pointer to an
// mgl32.Vec2 passed.
func Vec2(src *bytes.Buffer, x *mgl32.Vec2) error {
	return chainErr(
		Float32(src, &(*x)[0]),
		Float32(src, &(*x)[1]),
	)
}

// WriteVec2 writes an mgl32.Vec2 (float32 vector) to Buffer dst, writing each of the float32s separately.
func WriteVec2(dst *bytes.Buffer, x mgl32.Vec2) error {
	return chainErr(
		WriteFloat32(dst, x[0]),
		WriteFloat32(dst, x[1]),
	)
}
