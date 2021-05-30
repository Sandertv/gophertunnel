package packet

import (
	"bytes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// encrypt holds an encryption session with several fields required to encrypt and/or decrypt incoming
// packets. It may be initialised using secret key bytes computed using the shared secret produced with a
// private and a public ECDSA key.
type encrypt struct {
	sendCounter uint64
	buf         [8]byte
	keyBytes    []byte
	iv          []byte
	stream      cipher.AEAD
}

// newEncrypt returns a new encryption 'session' using the secret key bytes passed. The session has its cipher
// block and IV prepared so that it may be used to decrypt and encrypt data.
func newEncrypt(keyBytes []byte, iv []byte, stream cipher.AEAD) *encrypt {
	return &encrypt{keyBytes: keyBytes, iv: iv, stream: stream}
}

// encrypt encrypts the data passed, adding the packet checksum at the end of it before CFB8 encrypting it.
func (encrypt *encrypt) encrypt(data []byte) []byte {
	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	binary.LittleEndian.PutUint64(encrypt.buf[:], encrypt.sendCounter)
	encrypt.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(encrypt.buf[:])
	hash.Write(data[1:])
	hash.Write(encrypt.keyBytes)

	// We add the first 8 bytes of the checksum to the data and encrypt it.
	data = append(data, hash.Sum(nil)[:8]...)

	return append(data[0:1], encrypt.stream.Seal(data[1:1], encrypt.iv, data[1:], nil)...)
}

// decrypt decrypts the data passed. It does not verify the packet checksum. Verifying the checksum should be
// done using encrypt.verify(data).
func (encrypt *encrypt) decrypt(data []byte) error {
	if _, err := encrypt.stream.Open(data[:0], encrypt.iv, data, nil); err != nil {
		return err
	}
	return nil
}

// verify verifies the packet checksum of the decrypted data passed. If successful, nil is returned. Otherwise
// an error is returned describing the invalid checksum.
func (encrypt *encrypt) verify(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("encrypted packet must be at least 8 bytes long, got %v", len(data))
	}
	sum := data[len(data)-8:]

	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	binary.LittleEndian.PutUint64(encrypt.buf[:], encrypt.sendCounter)
	encrypt.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(encrypt.buf[:])
	hash.Write(data[:len(data)-8])
	hash.Write(encrypt.keyBytes)
	ourSum := hash.Sum(nil)[:8]

	// Finally we check if the original sum was equal to the sum we just produced.
	if !bytes.Equal(sum, ourSum) {
		return fmt.Errorf("invalid checksum of packet %v (%x): %x should be %x", encrypt.sendCounter-1, data, sum, ourSum)
	}
	return nil
}
