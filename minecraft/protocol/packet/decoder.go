package packet

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
	"io/ioutil"
)

// Decoder handles the decoding of Minecraft packets sent through an io.Reader. These packets in turn contain
// multiple zlib compressed packets.
type Decoder struct {
	buf    []byte
	reader io.Reader

	encrypt *encrypt
}

// NewDecoder returns a new decoder decoding data from the reader passed. One read call from the reader is
// assumed to consume an entire packet.
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader: reader,
		buf:    make([]byte, 1024*1024*3),
	}
}

// EnableEncryption enables encryption for the Decoder using the secret key bytes passed. Each packet received
// will be decrypted.
func (decoder *Decoder) EnableEncryption(keyBytes [32]byte) {
	decoder.encrypt = newEncrypt(keyBytes)
}

// header is the header of compressed 'batches' from Minecraft.
const header = 0xfe

// Decode decodes one 'packet' from the reader passed in NewDecoder(), producing a slice of packets that it
// held and an error if not successful.
func (decoder *Decoder) Decode() (packets [][]byte, err error) {
	n, err := decoder.reader.Read(decoder.buf)
	if err != nil {
		return nil, fmt.Errorf("error reading batch from reader: %v", err)
	}
	data := decoder.buf[:n]
	if data[0] != header {
		return nil, fmt.Errorf("error reading batch: invalid packet header %x: expected %x", data[0], header)
	}
	data = data[1:]
	if decoder.encrypt != nil {
		decoder.encrypt.decrypt(data)
		if err := decoder.encrypt.verify(data); err != nil {
			// The packet was not encrypted properly.
			return nil, fmt.Errorf("error reading batch: %v", err)
		}
	}

	raw, err := decoder.decompress(data)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(raw)
	for b.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(b, &length); err != nil {
			return nil, fmt.Errorf("error reading packet length: %v", err)
		}
		packets = append(packets, b.Next(int(length)))
	}
	return
}

// decompress zlib decompresses the data passed and returns it as a byte slice.
func (decoder *Decoder) decompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	zlibReader, err := zlib.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("error decompressing data: %v", err)
	}
	_ = zlibReader.Close()
	raw, err := ioutil.ReadAll(zlibReader)
	if err != nil {
		return nil, fmt.Errorf("error reading decompressed data: %v", err)
	}
	return raw, nil
}
