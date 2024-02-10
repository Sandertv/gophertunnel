package packet

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Decoder handles the decoding of Minecraft packets sent through an io.Reader. These packets in turn contain
// multiple compressed packets.
type Decoder struct {
	// r holds the io.Reader that packets are read from if the reader does not implement packetReader. When
	// this is the case, the buf field has a non-zero length.
	r   io.Reader
	buf []byte

	// pr holds a packetReader (and io.Reader) that packets are read from if the io.Reader passed to
	// NewDecoder implements the packetReader interface.
	pr packetReader

	decompress        bool
	compressionMethod Compression

	encryption Encryption

	checkPacketLimit bool
}

// packetReader is used to read packets immediately instead of copying them in a buffer first. This is a
// specific case made to reduce RAM usage.
type packetReader interface {
	ReadPacket() ([]byte, error)
}

// NewDecoder returns a new decoder decoding data from the io.Reader passed. One read call from the reader is
// assumed to consume an entire packet.
func NewDecoder(reader io.Reader) *Decoder {
	if pr, ok := reader.(packetReader); ok {
		return &Decoder{checkPacketLimit: true, pr: pr}
	}
	return &Decoder{
		r:                reader,
		buf:              make([]byte, 1024*1024*3),
		checkPacketLimit: true,
	}
}

// EnableEncryption enables encryption for the Decoder using the secret key bytes passed. Each packet received
// will be decrypted.
func (decoder *Decoder) EnableEncryption(encryption Encryption) {
	decoder.encryption = encryption
}

// EnableCompression enables compression for the Decoder.
func (decoder *Decoder) EnableCompression() {
	decoder.decompress = true
}

// SetCompression sets the compression method to use for the Decoder. This method should be used for versions below 1.20.60.
func (decoder *Decoder) SetCompression(method Compression) {
	decoder.compressionMethod = method
}

// DisableBatchPacketLimit disables the check that limits the number of packets allowed in a single packet
// batch. This should typically be called for Decoders decoding from a server connection.
func (decoder *Decoder) DisableBatchPacketLimit() {
	decoder.checkPacketLimit = false
}

const (
	// header is the header of compressed 'batches' from Minecraft.
	header = 0xfe
	// maximumInBatch is the maximum amount of packets that may be found in a batch. If a compressed batch has
	// more than this amount, decoding will fail.
	maximumInBatch = 812
)

// Decode decodes one 'packet' from the io.Reader passed in NewDecoder(), producing a slice of packets that it
// held and an error if not successful.
func (decoder *Decoder) Decode() (packets [][]byte, err error) {
	var data []byte
	if decoder.pr == nil {
		var n int
		n, err = decoder.r.Read(decoder.buf)
		data = decoder.buf[:n]
	} else {
		data, err = decoder.pr.ReadPacket()
	}
	if err != nil {
		return nil, fmt.Errorf("error reading batch from reader: %v", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	if data[0] != header {
		return nil, fmt.Errorf("error reading packet: invalid packet header %x: expected %x", data[0], header)
	}
	data = data[1:]
	if decoder.encryption != nil {
		decoder.encryption.Decrypt(data)
		if err := decoder.encryption.Verify(data); err != nil {
			// The packet did not have a correct checksum.
			return nil, fmt.Errorf("error verifying packet: %v", err)
		}
		data = data[:len(data)-8]
	}

	if decoder.decompress {
		if data[0] == 0xff {
			data = data[1:]
		} else {
			compression, ok := CompressionByID(uint16(data[0]))
			if !ok {
				return nil, fmt.Errorf("error decompressing packet: unknown compression algorithm %v", data[0])
			}
			data, err = compression.Decompress(data[1:])
			if err != nil {
				return nil, fmt.Errorf("error decompressing packet: %v", err)
			}
		}
	} else if decoder.compressionMethod != nil {
		data, err = decoder.compressionMethod.Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("error decompressing packet: %v", err)
		}
	}

	b := bytes.NewBuffer(data)
	for b.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(b, &length); err != nil {
			return nil, fmt.Errorf("error reading packet length: %v", err)
		}
		packets = append(packets, b.Next(int(length)))
	}
	if len(packets) > maximumInBatch && decoder.checkPacketLimit {
		return nil, fmt.Errorf("number of packets %v in compressed batch exceeds %v", len(packets), maximumInBatch)
	}
	return packets, nil
}
