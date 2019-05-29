package protocol

import (
	"bytes"
)

// BlockPos is the position of a block. It is composed of three integers, and is typically written as either
// 3 varint32s or a varint32, varuint32 and varint32.
type BlockPos [3]int32

// X returns the X coordinate of the block position. It is equivalent to BlockPos[0].
func (pos BlockPos) X() int32 {
	return pos[0]
}

// Y returns the Y coordinate of the block position. It is equivalent to BlockPos[1].
func (pos BlockPos) Y() int32 {
	return pos[1]
}

// Z returns the Z coordinate of the block position. It is equivalent to BlockPos[2].
func (pos BlockPos) Z() int32 {
	return pos[2]
}

// BlockPosition reads a BlockPos from Buffer src and stores it to the BlockPos pointer passed.
func BlockPosition(src *bytes.Buffer, x *BlockPos) error {
	if err := chainErr(
		Varint32(src, &(*x)[0]),
		Varint32(src, &(*x)[1]),
		Varint32(src, &(*x)[2]),
	); err != nil {
		return wrap(err)
	}
	return nil
}

// WriteBlockPosition writes a BlockPos x to Buffer dst, composed of 3 varint32s.
func WriteBlockPosition(dst *bytes.Buffer, x BlockPos) error {
	return chainErr(
		WriteVarint32(dst, x[0]),
		WriteVarint32(dst, x[1]),
		WriteVarint32(dst, x[2]),
	)
}

// UBlockPosition reads an unsigned BlockPos from Buffer src and stores it to the BlockPos pointer passed. The
// difference between this and BlockPosition is that the Y coordinate is read as a varuint32.
func UBlockPosition(src *bytes.Buffer, x *BlockPos) error {
	var v uint32
	if err := chainErr(
		Varint32(src, &(*x)[0]),
		Varuint32(src, &v),
		Varint32(src, &(*x)[2]),
	); err != nil {
		return wrap(err)
	}
	(*x)[1] = int32(v)
	return nil
}

// WriteUBlockPosition writes an unsigned BlockPos x to Buffer dst, composed of a varint32, varuint32 and a
// varint32.
func WriteUBlockPosition(dst *bytes.Buffer, x BlockPos) error {
	return chainErr(
		WriteVarint32(dst, x[0]),
		WriteVaruint32(dst, uint32(x[1])),
		WriteVarint32(dst, x[2]),
	)
}
