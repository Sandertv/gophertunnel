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
		return err
	}
	*x = math.Float32frombits(bits)
	return nil
}

// WriteFloat32 writes a float32 to Buffer dst, by first converting it to a uint32.
func WriteFloat32(dst *bytes.Buffer, x float32) error {
	return binary.Write(dst, binary.LittleEndian, math.Float32bits(x))
}

// Vec3 reads an mgl32.Vec3 (float32 vector) from Buffer src, setting the result to the pointer to an
// mgl32.Vec3 passed.
func Vec3(src *bytes.Buffer, x *mgl32.Vec3) error {
	if err := Float32(src, &(*x)[0]); err != nil {
		return err
	}
	if err := Float32(src, &(*x)[1]); err != nil {
		return err
	}
	return Float32(src, &(*x)[2])
}

// WriteVec3 writes an mgl32.Vec3 (float32 vector) to Buffer dst, writing each of the float32s separately.
func WriteVec3(dst *bytes.Buffer, x mgl32.Vec3) error {
	if err := WriteFloat32(dst, x[0]); err != nil {
		return err
	}
	if err := WriteFloat32(dst, x[1]); err != nil {
		return err
	}
	return WriteFloat32(dst, x[2])
}
