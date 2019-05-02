package jwt

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// New produces an encoded JWT from the header and payload passed. The signature of the JWT is created using
// the private key passed.
func New(header Header, payload interface{}, privateKey *ecdsa.PrivateKey) (string, error) {
	// First JSON+base64 encode the header.
	headerRaw, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("error JSON encoding header: %v", err)
	}
	headerSection := base64.RawURLEncoding.EncodeToString(headerRaw)

	// After that, we JSON+base64 encode the payload.
	payloadRaw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error JSON encoding payload: %v", err)
	}
	payloadSection := base64.RawURLEncoding.EncodeToString(payloadRaw)

	// The data we need to sign with the private key is the header and the payload joined by a dot.
	dataToSign := []byte(headerSection + "." + payloadSection)
	checksum := sha512.New384()
	checksum.Write(dataToSign)

	// We produce a signature which exists out of an 'r' and an 's', which we join to create the full
	// signature of the JWT.
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, checksum.Sum(nil))
	if err != nil {
		return "", fmt.Errorf("error signing JWT payload: %v", err)
	}
	signatureSection := base64.RawURLEncoding.EncodeToString(append(r.Bytes(), s.Bytes()...))

	// Finally we join together all sections and return it as a single string.
	return headerSection + "." + payloadSection + "." + signatureSection, nil
}
