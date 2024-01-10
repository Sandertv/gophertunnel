package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Encryption represents an interface for encrypting, decrypting, and verifying batches of data.
type Encryption interface {
	// Encrypt encrypts the data passed, adding the packet checksum at the end of it before encrypting it.
	Encrypt(data []byte) []byte
	// Decrypt decrypts the data passed. It does not verify the packet checksum. Verifying the checksum should be
	// done using (Encryption).Verify(data).
	Decrypt(data []byte)
	// Verify verifies the packet checksum of the decrypted data passed. If successful, nil is returned. Otherwise,
	// an error is returned describing the invalid checksum.
	Verify(data []byte) error
}

// ctr holds an encryption session with several fields required to encryption and/or decrypt incoming
// packets. It may be initialised using secret key bytes computed using the shared secret produced with a
// private and a public ECDSA key.
type ctr struct {
	sendCounter uint64
	buf         [8]byte
	keyBytes    []byte
	stream      cipher.Stream
}

// NewCTREncryption returns a new CTR encryption 'session' using the secret key bytes passed. The session has its cipher
// block and IV prepared so that it may be used to decrypt and encryption data.
func NewCTREncryption(keyBytes []byte) Encryption {
	block, _ := aes.NewCipher(keyBytes[:])
	first12 := append([]byte(nil), keyBytes[:12]...)
	stream := cipher.NewCTR(block, append(first12, 0, 0, 0, 2))
	return &ctr{keyBytes: keyBytes, stream: stream}
}

// Encrypt ...
func (c *ctr) Encrypt(data []byte) []byte {
	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	binary.LittleEndian.PutUint64(c.buf[:], c.sendCounter)
	c.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(c.buf[:])
	hash.Write(data[1:])
	hash.Write(c.keyBytes)

	// We add the first 8 bytes of the checksum to the data and encryption it.
	data = append(data, hash.Sum(nil)[:8]...)

	c.stream.XORKeyStream(data[1:], data[1:])
	return data
}

// Decrypt ...
func (c *ctr) Decrypt(data []byte) {
	c.stream.XORKeyStream(data, data)
}

// Verify ...
func (c *ctr) Verify(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("encrypted packet must be at least 8 bytes long, got %v", len(data))
	}
	sum := data[len(data)-8:]

	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	binary.LittleEndian.PutUint64(c.buf[:], c.sendCounter)
	c.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(c.buf[:])
	hash.Write(data[:len(data)-8])
	hash.Write(c.keyBytes)
	ourSum := hash.Sum(nil)[:8]

	// Finally we check if the original sum was equal to the sum we just produced.
	if !bytes.Equal(sum, ourSum) {
		return fmt.Errorf("invalid checksum of packet %v (%x): %x should be %x", c.sendCounter-1, data, sum, ourSum)
	}
	return nil
}
