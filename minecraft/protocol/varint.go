package protocol

import (
	"errors"
	"io"
)

// Varint64 reads up to 10 bytes from the source buffer passed and sets the integer produced to a pointer.
func Varint64(src io.ByteReader, x *int64) error {
	var ux uint64
	if err := Varuint64(src, &ux); err != nil {
		return err
	}
	*x = int64(ux >> 1)
	if ux&1 != 0 {
		*x = ^*x
	}
	return nil
}

// Varuint64 reads up to 10 bytes from the source buffer passed and sets the integer produced to a pointer.
func Varuint64(src io.ByteReader, x *uint64) error {
	var v uint64
	for i := uint(0); i < 70; i += 7 {
		b, err := src.ReadByte()
		if err != nil {
			return err
		}
		v |= uint64(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return nil
		}
	}
	return errors.New("varuint64 did not terminate after 10 bytes")
}

// Varint32 reads up to 5 bytes from the source buffer passed and sets the integer produced to a pointer.
func Varint32(src io.ByteReader, x *int32) error {
	var ux uint32
	if err := Varuint32(src, &ux); err != nil {
		return err
	}
	*x = int32(ux >> 1)
	if ux&1 != 0 {
		*x = ^*x
	}
	return nil
}

// Varuint32 reads up to 5 bytes from the source buffer passed and sets the integer produced to a pointer.
func Varuint32(src io.ByteReader, x *uint32) error {
	var v uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := src.ReadByte()
		if err != nil {
			return err
		}
		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			*x = v
			return nil
		}
	}
	return errors.New("varuint32 did not terminate after 5 bytes")
}

// WriteVarint64 writes an int64 to the destination buffer passed with a size of 1-10 bytes.
func WriteVarint64(dst io.ByteWriter, x int64) error {
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return WriteVaruint64(dst, ux)
}

// WriteVaruint64 writes a uint64 to the destination buffer passed with a size of 1-10 bytes.
func WriteVaruint64(dst io.ByteWriter, x uint64) error {
	for x >= 0x80 {
		if err := dst.WriteByte(byte(x) | 0x80); err != nil {
			return err
		}
		x >>= 7
	}
	return dst.WriteByte(byte(x))
}

// WriteVarint32 writes an int32 to the destination buffer passed with a size of 1-5 bytes.
func WriteVarint32(dst io.ByteWriter, x int32) error {
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	return WriteVaruint32(dst, ux)
}

// WriteVaruint32 writes a uint32 to the destination buffer passed with a size of 1-5 bytes.
func WriteVaruint32(dst io.ByteWriter, x uint32) error {
	for x >= 0x80 {
		if err := dst.WriteByte(byte(x) | 0x80); err != nil {
			return err
		}
		x >>= 7
	}
	return dst.WriteByte(byte(x))
}
