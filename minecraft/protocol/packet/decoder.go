package packet

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
)

// Decoder handles the decoding of Minecraft packets sent through an io.Reader. These packets in turn contain
// multiple compressed packets.
type Decoder struct {
	buf          []byte
	decompressor io.ReadCloser
	reader       io.Reader
	packetReader packetReader

	encrypt *encrypt

	checkPacketLimit bool
}

// packetReader is used to read packets immediately instead of copying them in a buffer first. This is a
// specific case made to reduce RAM usage.
type packetReader interface {
	ReadPacket() ([]byte, error)
}

// NewDecoder returns a new decoder decoding data from the reader passed. One read call from the reader is
// assumed to consume an entire packet.
func NewDecoder(reader io.Reader) *Decoder {
	if pr, ok := reader.(packetReader); ok {
		return &Decoder{checkPacketLimit: true, packetReader: pr}
	}
	return &Decoder{
		reader:           reader,
		buf:              make([]byte, 1024*1024*3),
		checkPacketLimit: true,
	}
}

// EnableEncryption enables encryption for the Decoder using the secret key bytes passed. Each packet received
// will be decrypted.
func (decoder *Decoder) EnableEncryption(keyBytes [32]byte) {
	block, _ := aes.NewCipher(keyBytes[:])
	decoder.encrypt = newEncrypt(keyBytes[:], newCFB8Decrypter(block, keyBytes[:aes.BlockSize]))
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
	maximumInBatch = 512 + 256
)

// Decode decodes one 'packet' from the reader passed in NewDecoder(), producing a slice of packets that it
// held and an error if not successful.
func (decoder *Decoder) Decode() (packets [][]byte, err error) {
	var data []byte
	if decoder.packetReader == nil {
		var n int
		n, err = decoder.reader.Read(decoder.buf)
		data = decoder.buf[:n]
	} else {
		data, err = decoder.packetReader.ReadPacket()
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
	if decoder.encrypt != nil {
		decoder.encrypt.decrypt(data)
		if err := decoder.encrypt.verify(data); err != nil {
			// The packet was not encrypted properly.
			return nil, fmt.Errorf("error verifying packet: %v", err)
		}
	}

	b, err := decoder.decompress(data)
	if err != nil {
		return nil, err
	}
	for b.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(b, &length); err != nil {
			return nil, fmt.Errorf("error reading packet length: %v", err)
		}
		if length == 0 {
			continue
		}
		packets = append(packets, b.Next(int(length)))
	}
	if len(packets) > maximumInBatch && decoder.checkPacketLimit {
		return nil, fmt.Errorf("number of packets %v in compressed batch exceeds %v", len(packets), maximumInBatch)
	}
	return packets, nil
}

// decompress decompresses the data passed and returns it as a byte slice.
func (decoder *Decoder) decompress(data []byte) (*bytes.Buffer, error) {
	if err := decoder.init(bytes.NewBuffer(data)); err != nil {
		return nil, fmt.Errorf("error decompressing data: %v", err)
	}
	_ = decoder.decompressor.Close()

	raw := bytes.NewBuffer(make([]byte, 0, len(data)*2))
	if _, err := io.Copy(raw, decoder.decompressor); err != nil {
		return nil, fmt.Errorf("error reading decompressed data: %v", err)
	}
	return raw, nil
}

// init initialises the decompression reader if it wasn't already.
func (decoder *Decoder) init(buf *bytes.Buffer) (err error) {
	if decoder.decompressor == nil {
		decoder.decompressor = flate.NewReader(buf)
		return
	}
	// The reader was already initialised, so we reset it to the buffer passed.
	return decoder.decompressor.(flate.Resetter).Reset(buf, nil)
}
