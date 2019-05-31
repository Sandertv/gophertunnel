package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// encrypt holds an encryption session with several fields required to encrypt and/or decrypt incoming
// packets. It may be initialised using secret key bytes computed using the shared secret produced with a
// private and a public ECDSA key.
type encrypt struct {
	sendCounter int64
	keyBytes    [32]byte
	cipherBlock cipher.Block
	iv          []byte
}

// newEncrypt returns a new encryption 'session' using the secret key bytes passed. The session has its cipher
// block and IV prepared so that it may be used to decrypt and encrypt data.
func newEncrypt(keyBytes [32]byte) *encrypt {
	block, _ := aes.NewCipher(keyBytes[:])
	return &encrypt{
		keyBytes:    keyBytes,
		cipherBlock: block,
		iv:          append([]byte(nil), keyBytes[:aes.BlockSize]...),
	}
}

// encrypt encrypts the data passed, adding the packet checksum at the end of it before CFB8 encrypting it.
func (encrypt *encrypt) encrypt(data []byte) []byte {
	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, encrypt.sendCounter)
	encrypt.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(buf.Bytes()[:8])
	hash.Write(data[1:])
	hash.Write(encrypt.keyBytes[:])

	// We add the first 8 bytes of the checksum to the data and encrypt it.
	data = append(data, hash.Sum(nil)[:8]...)

	// We skip the very first byte as it contains the header which we need not to encrypt.
	for i := range data[:len(data)-1] {
		offset := i + 1
		// We have to create a new CFBEncrypter for each byte that we decrypt, as this is CFB8.
		encrypter := cipher.NewCFBEncrypter(encrypt.cipherBlock, encrypt.iv)
		encrypter.XORKeyStream(data[offset:offset+1], data[offset:offset+1])
		// For each byte we encrypt, we need to update the IV we have. Each byte encrypted is added to the end
		// of the IV so that the first byte of the IV 'falls off'.
		encrypt.iv = append(encrypt.iv[1:], data[offset])
	}
	return data
}

// decrypt decrypts the data passed. It does not verify the packet checksum. Verifying the checksum should be
// done using encrypt.verify(data).
func (encrypt *encrypt) decrypt(data []byte) {
	for offset, b := range data {
		// Create a new CFBDecrypter for each byte, as we're dealing with CFB8 and have a new IV after each
		// byte that we decrypt.
		decrypter := cipher.NewCFBDecrypter(encrypt.cipherBlock, encrypt.iv)
		decrypter.XORKeyStream(data[offset:offset+1], data[offset:offset+1])

		// Each byte that we decrypt should be added to the end of the IV so that the first byte 'falls off'.
		encrypt.iv = append(encrypt.iv[1:], b)
	}
}

// verify verifies the packet checksum of the decrypted data passed. If successful, nil is returned. Otherwise
// an error is returned describing the invalid checksum.
func (encrypt *encrypt) verify(data []byte) error {
	sum := data[len(data)-8:]

	// We first write the current send counter to a buffer and use it to produce a packet checksum.
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	_ = binary.Write(buf, binary.LittleEndian, encrypt.sendCounter)
	encrypt.sendCounter++

	// We produce a hash existing of the send counter, packet data and key bytes.
	hash := sha256.New()
	hash.Write(buf.Bytes())
	hash.Write(data[:len(data)-8])
	hash.Write(encrypt.keyBytes[:])
	ourSum := hash.Sum(nil)[:8]

	// Finally we check if the original sum was equal to the sum we just produced.
	if !bytes.Equal(sum, ourSum) {
		return fmt.Errorf("invalid packet checksum: %v should be %v", hex.EncodeToString(sum), hex.EncodeToString(ourSum))
	}
	return nil
}
