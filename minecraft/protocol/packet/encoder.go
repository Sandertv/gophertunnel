package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/internal"
	"io"
)

// Encoder handles the encoding of Minecraft packets that are sent to an io.Writer. The packets are compressed
// and optionally encoded before they are sent to the io.Writer.
type Encoder struct {
	w io.Writer

	compression Compression
	encrypt     *encrypt
	header      []byte
}

// NewEncoder returns a new Encoder for the io.Writer passed. Each final packet produced by the Encoder is
// sent with a single call to io.Writer.Write().
func NewEncoder(w io.Writer, header []byte) *Encoder {
	return &Encoder{w: w, header: header}
}

// EnableEncryption enables encryption for the Encoder using the secret key bytes passed. Each packet sent
// after encryption is enabled will be encrypted.
func (encoder *Encoder) EnableEncryption(keyBytes [32]byte) {
	block, _ := aes.NewCipher(keyBytes[:])
	first12 := append([]byte(nil), keyBytes[:12]...)
	stream := cipher.NewCTR(block, append(first12, 0, 0, 0, 2))
	encoder.encrypt = newEncrypt(keyBytes[:], stream)
}

// EnableCompression enables compression for the Encoder.
func (encoder *Encoder) EnableCompression(compression Compression) {
	encoder.compression = compression
}

// Encode encodes the packets passed. It writes all of them as a single packet which is  compressed and
// optionally encrypted.
func (encoder *Encoder) Encode(packets [][]byte) error {
	buf := internal.BufferPool.Get().(*bytes.Buffer)
	defer func() {
		// Reset the buffer, so we can return it to the buffer pool safely.
		buf.Reset()
		internal.BufferPool.Put(buf)
	}()

	l := make([]byte, 5)
	for _, packet := range packets {
		// Each packet is prefixed with a varuint32 specifying the length of the packet.
		if err := writeVaruint32(buf, uint32(len(packet)), l); err != nil {
			return fmt.Errorf("encode batch: write packet length: %w", err)
		}
		if _, err := buf.Write(packet); err != nil {
			return fmt.Errorf("encode batch: write packet payload: %w", err)
		}
	}

	data := buf.Bytes()
	prepend := encoder.header
	if encoder.compression != nil {
		prepend = append(prepend, byte(encoder.compression.EncodeCompression()))
		var err error
		data, err = encoder.compression.Compress(data)
		if err != nil {
			return fmt.Errorf("compress batch: %w", err)
		}
	}

	data = append(prepend, data...)
	if encoder.encrypt != nil {
		// If the encryption session is not nil, encryption is enabled, meaning we should encrypt the
		// compressed data of this packet.
		data = encoder.encrypt.encrypt(data)
	}
	if _, err := encoder.w.Write(data); err != nil {
		return fmt.Errorf("write batch: %w", err)
	}
	return nil
}

// writeVaruint32 writes a uint32 to the destination buffer passed with a size of 1-5 bytes. It uses byte
// slice b in order to prevent allocations.
func writeVaruint32(dst io.Writer, x uint32, b []byte) error {
	clear(b[:5])
	i := 0
	for x >= 0x80 {
		b[i] = byte(x) | 0x80
		i++
		x >>= 7
	}
	b[i] = byte(x)
	_, err := dst.Write(b[:i+1])
	return err
}
