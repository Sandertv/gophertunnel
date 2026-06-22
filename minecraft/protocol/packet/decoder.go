package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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

	// pr holds a PacketReader (and io.Reader) that packets are read from if the io.Reader passed to
	// NewDecoder implements the PacketReader interface.
	pr PacketReader

	// header holds the batch header that is expected on the beginning of input packet data.
	header             []byte
	decompress         bool
	compression        Compression
	maxDecompressedLen int
	encrypt            *encrypt
	// disableEncryption indicates whether to prevent encryption from being enabled
	// even if it is requested on handshake during login.
	disableEncryption bool

	checkPacketLimit bool
}

// NewDecoder returns a new decoder decoding data from the io.Reader passed. One read call from the reader is
// assumed to consume an entire packet.
func NewDecoder(reader io.Reader) *Decoder {
	var batch []byte
	if b, ok := reader.(BatchHeaderer); ok {
		batch = b.BatchHeader()
	} else {
		batch = []byte{header}
	}
	var disableEncryption bool
	if d, ok := reader.(EncryptionDisabler); ok {
		disableEncryption = d.DisableEncryption()
	}
	if pr, ok := reader.(PacketReader); ok {
		return &Decoder{
			checkPacketLimit:  true,
			pr:                pr,
			header:            batch,
			disableEncryption: disableEncryption,
		}
	}
	return &Decoder{
		r:                 reader,
		buf:               make([]byte, 1024*1024*3),
		header:            batch,
		checkPacketLimit:  true,
		disableEncryption: disableEncryption,
	}
}

// EnableEncryption enables encryption for the Decoder using the secret key bytes passed. Each packet received
// will be decrypted.
func (decoder *Decoder) EnableEncryption(keyBytes [32]byte) {
	if decoder.disableEncryption {
		return
	}
	block, _ := aes.NewCipher(keyBytes[:])
	first12 := append([]byte(nil), keyBytes[:12]...)
	stream := cipher.NewCTR(block, append(first12, 0, 0, 0, 2))
	decoder.encrypt = newEncrypt(keyBytes[:], stream)
}

// EnableCompression enables compression for the Decoder.
func (decoder *Decoder) EnableCompression(compression Compression, maxDecompressedLen int) {
	decoder.decompress = true
	decoder.compression = compression
	decoder.maxDecompressedLen = maxDecompressedLen
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
	maximumInBatch = 1600
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
		return nil, fmt.Errorf("read batch: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}
	h := len(decoder.header)
	if len(data) < h {
		return nil, io.ErrUnexpectedEOF
	}
	if !bytes.Equal(data[:h], decoder.header) {
		return nil, fmt.Errorf("decode batch: invalid header %x, expected %x", data[:h], decoder.header)
	}
	data = data[h:]
	if decoder.encrypt != nil {
		decoder.encrypt.decrypt(data)
		if err := decoder.encrypt.verify(data); err != nil {
			// The packet did not have a correct checksum.
			return nil, fmt.Errorf("verify batch: %w", err)
		}
		data = data[:len(data)-8]
	}

	if decoder.decompress {
		if len(data) == 0 {
			return nil, fmt.Errorf("decompress batch: missing compression algorithm")
		}
		if data[0] == 0xff {
			data = data[1:]
		} else {
			compression, ok := CompressionByID(uint16(data[0]))
			if !ok {
				return nil, fmt.Errorf("decompress batch: unknown compression algorithm %v", data[0])
			}
			if compression != decoder.compression {
				return nil, fmt.Errorf("decompress batch: unexpected compression algorithm: got %v, expected %v", compression, decoder.compression)
			}
			data, err = compression.Decompress(data[1:], decoder.maxDecompressedLen)
			if err != nil {
				return nil, fmt.Errorf("decompress batch: %w", err)
			}
		}
	}

	b := bytes.NewBuffer(data)
	for b.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(b, &length); err != nil {
			return nil, fmt.Errorf("decode batch: read packet length: %w", err)
		}
		if length == 0 {
			return nil, fmt.Errorf("decode batch: empty packet")
		}
		if length > uint32(b.Len()) {
			return nil, fmt.Errorf("decode batch: packet length %v exceeds remaining %v", length, b.Len())
		}
		if len(packets) >= maximumInBatch && decoder.checkPacketLimit {
			return nil, fmt.Errorf("decode batch: number of packets exceeds max=%v", maximumInBatch)
		}
		packets = append(packets, b.Next(int(length)))
	}
	return packets, nil
}
