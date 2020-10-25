package packet

import (
	"crypto/aes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"github.com/sandertv/gophertunnel/internal/dynamic"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
)

// Encoder handles the encoding of Minecraft packets that are sent to an io.Writer. The packets are compressed
// and optionally encoded before they are sent to the io.Writer.
type Encoder struct {
	compressor      writeCloseResetter
	writer          io.Writer
	buf, compressed *dynamic.Buffer

	encrypt *encrypt
}

// writeCloseResetter is an interface composed of an io.WriteCloser and a Reset(io.Writer) method.
type writeCloseResetter interface {
	io.WriteCloser
	Reset(w io.Writer)
}

// NewEncoder returns a new Encoder for the io.Writer passed. Each final packet produced by the Encoder is
// sent with a single call to io.Writer.Write().
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		writer:     w,
		compressor: flate.NewStatelessWriter(w).(writeCloseResetter),
		buf:        dynamic.NewBuffer(make([]byte, 0, 1024*1024)),
		compressed: dynamic.NewBuffer(make([]byte, 0, 1024*1024)),
	}
}

// EnableEncryption enables encryption for the Encoder using the secret key bytes passed. Each packet sent
// after encryption is enabled will be encrypted.
func (encoder *Encoder) EnableEncryption(keyBytes [32]byte) {
	block, _ := aes.NewCipher(append([]byte(nil), keyBytes[:]...))
	encoder.encrypt = newEncrypt(append([]byte(nil), keyBytes[:]...), newCFB8Encrypter(block, append([]byte(nil), keyBytes[:aes.BlockSize]...)))
}

// Encode encodes the packets passed. It writes all of them as a single packet which is  compressed and
// optionally encrypted.
func (encoder *Encoder) Encode(packets [][]byte) error {
	defer func() {
		// Reset both buffers so that they can be re-used the next time Encoder encodes packets.
		encoder.buf.Reset()
		encoder.compressed.Reset()
	}()
	if err := encoder.buf.WriteByte(header); err != nil {
		return fmt.Errorf("error writing 0xfe header: %v", err)
	}

	for _, packet := range packets {
		// Each packet is prefixed with a varuint32 specifying the length of the packet.
		if err := protocol.WriteVaruint32(encoder.compressed, uint32(len(packet))); err != nil {
			return fmt.Errorf("error writing varuint32 length: %v", err)
		}
		if _, err := encoder.compressed.Write(packet); err != nil {
			return fmt.Errorf("error writing packet payload: %v", err)
		}
	}

	// We compress the data and write the full data to the io.Writer. The data returned includes the header
	// we wrote at the start.
	b, err := encoder.compress(encoder.compressor, encoder.compressed.Bytes())
	if err != nil {
		return err
	}

	if encoder.encrypt != nil {
		// If the encryption session is not nil, encryption is enabled, meaning we should encrypt the
		// compressed data of this packet.
		b = encoder.encrypt.encrypt(b)
	}
	if _, err := encoder.writer.Write(b); err != nil {
		return fmt.Errorf("error writing compressed packet to io.Writer: %v", err)
	}
	return nil
}

// compress compresses the data passed using the writer passed and returns it in a byte slice. It returns
// the full content of encoder.buf, so any data currently set in that buffer will also be returned.
func (encoder *Encoder) compress(w writeCloseResetter, data []byte) ([]byte, error) {
	w.Reset(encoder.buf)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("error writing compressed data: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error closing compressor: %v", err)
	}
	return encoder.buf.Bytes(), nil
}
