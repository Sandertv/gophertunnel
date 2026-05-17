package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"slices"

	"github.com/sandertv/gophertunnel/minecraft/internal"
)

// Encoder handles the encoding of Minecraft packets that are sent to an io.Writer. The packets are compressed
// and optionally encoded before they are sent to the io.Writer.
type Encoder struct {
	w io.Writer

	// header holds the batch header that should be present at the beginning on each produced packet data
	// held in a single batch packet.
	header               []byte
	compressionThreshold int
	compression          Compression
	encrypt              *encrypt
	// disableEncryption indicates whether to prevent encryption from being enabled
	// even if it is requested on handshake during login.
	disableEncryption bool
}

// NewEncoder returns a new Encoder for the io.Writer passed. Each final packet produced by the Encoder is
// sent with a single call to io.Writer.Write().
func NewEncoder(w io.Writer) *Encoder {
	var batch []byte
	if b, ok := w.(batchHeader); ok {
		batch = b.BatchHeader()
	} else {
		batch = []byte{header}
	}
	var disableEncryption bool
	if d, ok := w.(encryptionDisabler); ok {
		disableEncryption = d.DisableEncryption()
	}
	return &Encoder{
		w:                 w,
		header:            batch,
		disableEncryption: disableEncryption,
	}
}

// batchHeader can be implemented by underlying transport connection provided to Encoder and Decoder
// to specify the initial bytes that should appear at the beginning of packet data in wire.
type batchHeader interface {
	// BatchHeader returns initial bytes that should be appended to the produced data
	// in Encoder and Decoder. It can be an empty slice if nothing is expected at the beginning.
	BatchHeader() []byte
}

// encryptionDisabler may be implemented by the underlying transport connection to
// prevent encryption from being enabled in Encoder and Decoder.
//
// Disabling encryption is strongly discouraged, as it removes protection against
// replay attacks during login. Use only if you fully understand the implications.
type encryptionDisabler interface {
	// DisableEncryption reports whether encryption should be disabled for both
	// Encoder and Decoder.
	DisableEncryption() bool
}

// EnableEncryption enables encryption for the Encoder using the secret key bytes passed. Each packet sent
// after encryption is enabled will be encrypted.
func (encoder *Encoder) EnableEncryption(keyBytes [32]byte) {
	if encoder.disableEncryption {
		return
	}
	block, _ := aes.NewCipher(keyBytes[:])
	first12 := append([]byte(nil), keyBytes[:12]...)
	stream := cipher.NewCTR(block, append(first12, 0, 0, 0, 2))
	encoder.encrypt = newEncrypt(keyBytes[:], stream)
}

// EnableCompression enables compression for the Encoder.
func (encoder *Encoder) EnableCompression(compression Compression, threshold int) {
	encoder.compression = compression
	encoder.compressionThreshold = threshold
}

// Encode encodes the packets passed. It writes all of them as a single packet which is  compressed and
// optionally encrypted.
func (encoder *Encoder) Encode(packets [][]byte) error {
	buf := internal.BufferPool.Get().(*bytes.Buffer)
	var compressedBuf *bytes.Buffer
	defer func() {
		// Reset the buffer, so we can return it to the buffer pool safely.
		buf.Reset()
		internal.BufferPool.Put(buf)
		if compressedBuf != nil {
			compressedBuf.Reset()
			internal.BufferPool.Put(compressedBuf)
		}
	}()

	compression := encoder.compression
	_, _ = buf.Write(encoder.header)
	if compression != nil {
		_ = buf.WriteByte(0)
	}
	batchStart := buf.Len()

	var l [5]byte
	for _, packet := range packets {
		// Each packet is prefixed with a varuint32 specifying the length of the packet.
		if _, err := buf.Write(l[:putVaruint32(&l, uint32(len(packet)))]); err != nil {
			return fmt.Errorf("encode batch: write packet length: %w", err)
		}
		if _, err := buf.Write(packet); err != nil {
			return fmt.Errorf("encode batch: write packet payload: %w", err)
		}
	}

	data := buf.Bytes()
	if compression != nil {
		batch := data[batchStart:]
		if len(batch) < encoder.compressionThreshold {
			data[len(encoder.header)] = byte(NopCompression.EncodeCompression())
		} else {
			data[len(encoder.header)] = byte(compression.EncodeCompression())
			compressedBuf = internal.BufferPool.Get().(*bytes.Buffer)
			_, _ = compressedBuf.Write(encoder.header)
			_ = compressedBuf.WriteByte(byte(compression.EncodeCompression()))
			var err error
			if appender, ok := compression.(appendCompression); ok {
				if n := appender.MaxCompressedLen(len(batch)); n > 0 {
					compressedBuf.Grow(n)
				}
				dst := compressedBuf.Bytes()
				data, err = appender.CompressAppend(dst, batch)
			} else {
				dst := compressedBuf.Bytes()
				var compressed []byte
				compressed, err = compression.Compress(batch)
				data = append(dst, compressed...)
			}
			if err != nil {
				return fmt.Errorf("compress batch: %w", err)
			}
		}
	}

	if encoder.encrypt != nil {
		// If the encryption session is not nil, encryption is enabled, meaning we should encrypt the
		// compressed data of this packet.
		data = slices.Grow(data, 8)
		data = encoder.encrypt.encrypt(data)
	}
	if _, err := encoder.w.Write(data); err != nil {
		return fmt.Errorf("write batch: %w", err)
	}
	return nil
}

// putVaruint32 writes x to b with a size of 1-5 bytes and returns the number of
// bytes written.
func putVaruint32(b *[5]byte, x uint32) int {
	i := 0
	for x >= 0x80 {
		b[i] = byte(x) | 0x80
		i++
		x >>= 7
	}
	b[i] = byte(x)
	return i + 1
}
