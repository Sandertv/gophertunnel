package jwt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

//noinspection SpellCheckingInspection
const (
	// MojangPublicKey is the public key used by Mojang to sign one of the claims in a chain, indicating that
	// the player was logged into XBOX Live.
	MojangPublicKey = `MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE8ELkixyLcwlZryUQcu1TvPOmI2B7vX83ndnWRUaXm74wFfa5f/lwQNTfrLVHa2PmenpGI6JhIMUJaWZrjmMj90NoKNFSNBuKdm8rYiXsfaz3K36x/1U26HpG0ZxK/V1V`
)

// Verify verifies a single raw JWT string, which exists out of a header, payload and a signature. The JWT
// is first checked to be valid, after which its signature is verified.
// The publicKey passed is used to verify the signature of the claim. If a zero public key is passed (meaning,
// not a nil pointer, but an empty *ecdsa.PublicKey{}), the key is retrieved from the x5u of the header.
// The public key passed will be updated for the identityPublicKey found in the claim.
func Verify(jwt []byte, publicKey *ecdsa.PublicKey, needNewKey bool) (hasMojangKey bool, err error) {
	fragments := bytes.Split(jwt, []byte{'.'})
	if len(fragments) != 3 {
		return false, fmt.Errorf("expected claim to have 3 sections, but got %v", len(fragments))
	}
	for index, f := range fragments {
		// First base64 decode all of these fragments so we can directly assign them without having to decode
		// them one by one.
		// Some (faulty) JWT implementations use padded base64, whereas it should be raw. We trim this off.
		f = bytes.TrimRight(f, "=")
		b, err := base64.RawURLEncoding.DecodeString(string(f))
		if err != nil {
			return false, fmt.Errorf("error base64 decoding claim: %v", err)
		}
		fragments[index] = b
	}
	rawHeader, rawPayload, rawSignature := fragments[0], fragments[1], fragments[2]

	// Header validation.
	header := &Header{}
	if err := json.Unmarshal(rawHeader, header); err != nil {
		return false, fmt.Errorf("error decoding header: %v", err)
	}
	if publicKey.Y == nil {
		if err := ParsePublicKey(header.X5U, publicKey); err != nil {
			return false, fmt.Errorf("error parsing x5u: %v", err)
		}
	}
	if !AllowedAlg(header.Algorithm) {
		// The algorithm wasn't found in one of the allowed algorithms, so we return an error immediately and
		// stop verification.
		return false, fmt.Errorf("disallowed header algorithm %v: expected one of %v", header.Algorithm, allowedAlgorithms)
	}

	type jwtPayload struct {
		Exp               int64  `json:"exp"`
		Nbf               int64  `json:"nbf"`
		Iat               int64  `json:"iat"`
		IdentityPublicKey string `json:"identityPublicKey"`
	}

	// Payload validation.
	jwtData := jwtPayload{}
	if err := json.Unmarshal(rawPayload, &jwtData); err != nil {
		return false, fmt.Errorf("error decoding payload: %v", err)
	}
	// Account for one hour of clock drift.
	now := time.Now()
	if jwtData.Exp != 0 {
		if time.Unix(jwtData.Exp, 0).Before(now) {
			// The expiration time was before 'now', meaning the token was no longer usable.
			return false, fmt.Errorf("JWT claim expired: token is no longer usable")
		}
	}
	now = now.Add(time.Hour)
	if now.Before(time.Unix(jwtData.Nbf, 0)) {
		// The 'not before' or 'issued at' times were after now, meaning we shouldn't have possibly
		// been able to receive the token yet.
		return false, fmt.Errorf("JWT claim used too early: token is not yet usable")
	}
	if now.Before(time.Unix(jwtData.Iat, 0)) {
		// The 'not before' or 'issued at' times were after now, meaning we shouldn't have possibly
		// been able to receive the token yet.
		return false, fmt.Errorf("JWT claim used too early: token is not yet usable")
	}
	var newPublicKeyData string
	if needNewKey {
		newPublicKeyData = jwtData.IdentityPublicKey
		if newPublicKeyData == MojangPublicKey {
			hasMojangKey = true
		}
	}
	// Signature verification.
	hash := sha512.New384()
	// The hash is produced using the header and the payload section of the claim.
	hash.Write(jwt[:bytes.LastIndex(jwt, []byte{'.'})])

	sigLength := len(rawSignature)
	r := new(big.Int).SetBytes(bytes.TrimLeft(rawSignature[:sigLength/2], "\x00"))
	s := new(big.Int).SetBytes(bytes.TrimLeft(rawSignature[sigLength/2:], "\x00"))

	if !ecdsa.Verify(publicKey, hash.Sum(nil), r, s) {
		return false, fmt.Errorf("JWT claim has an incorrect signature")
	}

	if needNewKey {
		// Finally parse the new identityPublicKey and set it to the public key pointer passed, so that it may
		// be used to verify the next claim in the chain.
		if err := ParsePublicKey(newPublicKeyData, publicKey); err != nil {
			return false, fmt.Errorf("error parsing identityPublicKey: %v", err)
		}
	}
	return hasMojangKey, nil
}

// Payload parses the JWT passed and returns the base64 decoded payload section of the claim. The JSON data
// returned is not guaranteed to be valid JSON.
func Payload(jwt []byte) ([]byte, error) {
	fragments := bytes.Split(jwt, []byte{'.'})
	if len(fragments) != 3 {
		return nil, fmt.Errorf("expected claim to have 3 sections, but got %v", len(fragments))
	}
	// Some (faulty) JWT implementations use padded base64, whereas it should be raw. We trim this off.
	fragments[1] = bytes.TrimRight(fragments[1], "=")
	payload, err := base64.RawURLEncoding.DecodeString(string(fragments[1]))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding payload: %v", err)
	}
	return payload, nil
}

// ParsePublicKey parses a public key from the base64 encoded public key data passed and sets it to the public
// key pointer. If parsing failed or if the public key was not of the type ECDSA, an error is returned.
func ParsePublicKey(b64Data string, key *ecdsa.PublicKey) error {
	data, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return fmt.Errorf("error base64 decoding public key data: %v", err)
	}
	publicKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return fmt.Errorf("error parsing public key: %v", err)
	}
	ecdsaKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, but got %v", key)
	}
	*key = *ecdsaKey
	return nil
}

// MarshalPublicKey marshals an ECDSA public key to a base64 encoded binary representation.
func MarshalPublicKey(key *ecdsa.PublicKey) string {
	data, _ := x509.MarshalPKIXPublicKey(key)
	return base64.StdEncoding.EncodeToString(data)
}
