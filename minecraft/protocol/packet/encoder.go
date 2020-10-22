package packet

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
)

// Encoder handles the encoding of Minecraft packets that are sent to an io.Writer. The packets are compressed
// and optionally encoded before they are sent to the io.Writer.
type Encoder struct {
	compressor, mediumCompressor, noCompression *flate.Writer
	writer                                      io.Writer
	buf, compressed                             *bytes.Buffer

	bufShrinkTick        int
	compressedShrinkTick int

	encrypt *encrypt
}

// NewEncoder returns a new Encoder for the io.Writer passed. Each final packet produced by the Encoder is
// sent with a single call to io.Writer.Write().
func NewEncoder(writer io.Writer) *Encoder {
	w, _ := flate.NewWriter(writer, flate.NoCompression)
	w2, _ := flate.NewWriter(writer, flate.DefaultCompression)
	w3, _ := flate.NewWriter(writer, 4)
	return &Encoder{
		noCompression:    w,
		compressor:       w2,
		mediumCompressor: w3,
		writer:           writer,
		buf:              bytes.NewBuffer(make([]byte, 0, 1024*1024*2)),
		compressed:       bytes.NewBuffer(make([]byte, 0, 1024*1024*3)),
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

	var w *flate.Writer
	if encoder.compressed.Len() <= 512 {
		w = encoder.noCompression
	} else if encoder.compressed.Len() <= 8192 {
		w = encoder.mediumCompressor
	} else {
		w = encoder.compressor
	}
	// We compress the data and write the full data to the io.Writer. The data returned includes the header
	// we wrote at the start.
	b, err := encoder.compress(w, encoder.compressed.Bytes())
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

	encoder.shrinkBuffers()
	return nil
}

// shrinkBuffers shrinks the bytes.Buffers held by the Encoder when incoming data has consistently been
// considerably smaller than the size of the buffer.
func (encoder *Encoder) shrinkBuffers() {
	if encoder.compressed.Cap() > encoder.compressed.Len()*2 {
		encoder.compressedShrinkTick++
		if encoder.compressedShrinkTick > 40 {
			encoder.compressed = bytes.NewBuffer(make([]byte, 0, encoder.compressed.Len()/2))
		}
	} else if encoder.compressedShrinkTick > 0 {
		encoder.compressedShrinkTick--
	}
	if encoder.buf.Cap() > encoder.buf.Len()*2 {
		encoder.bufShrinkTick++
		if encoder.bufShrinkTick > 40 {
			encoder.buf = bytes.NewBuffer(make([]byte, 0, encoder.buf.Len()/2))
		}
	} else if encoder.bufShrinkTick > 0 {
		encoder.bufShrinkTick--
	}
}

// compress compresses the data passed using the writer passed and returns it in a byte slice. It returns
// the full content of encoder.buf, so any data currently set in that buffer will also be returned.
func (encoder *Encoder) compress(w *flate.Writer, data []byte) ([]byte, error) {
	w.Reset(encoder.buf)
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("error writing compressed data: %v", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error closing compressor: %v", err)
	}
	return encoder.buf.Bytes(), nil
}
