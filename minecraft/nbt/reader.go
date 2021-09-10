package nbt

import (
	"io"
)

// offsetReader is a wrapper around an io.Reader, used to track the offset (amount of bytes read) of the data
// that is being read, so that errors may have offset data.
type offsetReader struct {
	io.Reader
	off int64

	// ReadByte is a function provided by offsetReader if the io.Reader does not implement io.ByteReader.
	ReadByte func() (byte, error)
	// Next is a function provided by offsetReader if the io.Reader does not have a Next method.
	Next func(n int) []byte
}

// newOffsetReader returns a new offset reader for the io.Reader passed, setting the ReadByte and Next
// functions as appropriate for that particular reader.
func newOffsetReader(r io.Reader) *offsetReader {
	reader := &offsetReader{Reader: r}
	if byteReader, ok := r.(io.ByteReader); ok {
		reader.ReadByte = func() (byte, error) {
			reader.off++
			return byteReader.ReadByte()
		}
	} else {
		reader.ReadByte = func() (byte, error) {
			data := make([]byte, 1)
			_, err := io.ReadAtLeast(reader, data, 1)
			return data[0], err
		}
	}
	if r, ok := r.(interface {
		Next(n int) []byte
	}); ok {
		reader.Next = func(n int) []byte {
			data := r.Next(n)
			reader.off += int64(len(data))
			return data
		}
	} else {
		reader.Next = func(n int) []byte {
			data := make([]byte, n)
			_, _ = io.ReadAtLeast(reader, data, n)
			return data
		}
	}
	return reader
}

// Read reads from the io.Reader and increases the reader's offset by exactly n.
func (b *offsetReader) Read(p []byte) (n int, err error) {
	n, err = io.ReadAtLeast(b.Reader, p, len(p))
	b.off += int64(n)
	return
}
