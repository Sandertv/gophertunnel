package chat

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/service/internal"
)

// Verify attempts to verify the given chat message received from other players through the server.
// It returns nil if the verification was successful, otherwise a non-nil error.
// Callers can use errors.Is(err, ErrInvalidSignature) to check whether signature
// verification failed.
func (e *Environment) Verify(ctx context.Context, chat protocol.SignedChat) error {
	if chat.RoomID == "" || chat.KeyID == "" || chat.Signature == "" {
		return errors.New("service/chat: payload is not signed")
	}
	signature, err := base64.StdEncoding.DecodeString(chat.Signature)
	if err != nil {
		return fmt.Errorf("service/chat: decode signature: %w", err)
	}
	key, err := e.verificationKey(ctx, chat.KeyID)
	if err != nil {
		return fmt.Errorf("service/chat: lookup signing key: %w", err)
	}

	h := key.hash()
	writeString(h, chat.KeyID)
	writeString(h, chat.SenderID)
	writeString(h, chat.RoomID)
	writeString(h, chat.MessageID)
	writeString(h, chat.ScanText)
	writeString(h, chat.RedactedText)

	if !ecdsa.VerifyASN1(key.PublicKey, h.Sum(nil), signature) {
		return ErrInvalidSignature
	}
	return nil
}

// ErrInvalidSignature is the error returned by [Environment.Verify] when the chat signature
// cannot be verified using the public key retrieved from the Chat service.
var ErrInvalidSignature = errors.New("service/chat: invalid signature")

// writeString writes s prefixed by a big-endian uint32 length to w.
func writeString(w io.Writer, s string) {
	_ = binary.Write(w, binary.BigEndian, uint32(len(s)))
	_, _ = w.Write([]byte(s))
}

// verificationKey looks up for the [ecdsa.PublicKey] associated with the given key ID.
// It retrieves the entire key set if the cache is empty or has expired.
func (e *Environment) verificationKey(ctx context.Context, keyID string) (*verificationKey, error) {
	e.keysMu.Lock()
	defer e.keysMu.Unlock()

	if e.keys == nil || time.Now().After(e.keysExpiry) {
		keys, err := e.verificationKeys(ctx)
		if err != nil {
			return nil, fmt.Errorf("request signing keys: %w", err)
		}
		// TODO: Figure out the actual expiry on the client
		e.keys, e.keysExpiry = keys, time.Now().Add(time.Minute*15)
	}
	key, ok := e.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("signing key not found for %q", keyID)
	}
	return key, nil
}

// verificationKeys retrieves the set of public keys used to verify signatures
// produced by the Chat service.
func (e *Environment) verificationKeys(ctx context.Context) (map[string]*verificationKey, error) {
	requestURL := e.ServiceURI.JoinPath("/.well-known/signing").String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	req.Header.Set("User-Agent", internal.UserAgent)

	resp, err := e.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result struct {
			Keys []*verificationKey `json:"keys"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}
		if len(result.Keys) == 0 {
			return nil, errors.New("invalid signing keys response")
		}
		keys := make(map[string]*verificationKey)
		for _, key := range result.Keys {
			if key == nil {
				return nil, errors.New("key is nil") // should not happen
			}
			if err := key.validate(); err != nil {
				return nil, fmt.Errorf("validate key: %w", err)
			}
			keys[key.KeyID] = key
		}
		return keys, nil
	default:
		return nil, internal.Err(resp)
	}
}

// verificationKey represents a key used by the service to sign chat messages.
type verificationKey struct {
	// KeyID uniquely identifies the signing key.
	KeyID string `json:"kid"`
	// PublicKey is the public key encoded as a base64-encoded DER value.
	PublicKey *ecdsa.PublicKey `json:"publicKey"`
	// Algorithm specifies th signature algorithm used with this key.
	// It is typically 'ES256'.
	Algorithm string `json:"algorithm"`
}

// UnmarshalJSON implements [json.Unmarshaler] for verificationKey.
func (k *verificationKey) UnmarshalJSON(b []byte) error {
	type Alias verificationKey
	data := struct {
		*Alias
		PublicKey string `json:"publicKey"`
	}{Alias: (*Alias)(k)}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	k.PublicKey = new(ecdsa.PublicKey)
	if err := login.ParsePublicKey(data.PublicKey, k.PublicKey); err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	return nil
}

// validate validates whether verificationKey can be safe to use with signature verification.
func (k *verificationKey) validate() error {
	if k.KeyID == "" || k.PublicKey == nil || k.Algorithm == "" {
		return fmt.Errorf("insufficient fields")
	}
	// TODO: Support more signature algorithms in the future.
	var valid bool
	switch k.Algorithm {
	case "ES256":
		valid = k.PublicKey.Curve == elliptic.P256()
	default:
		return fmt.Errorf("unknown algorithm: %q", k.Algorithm)
	}
	if !valid {
		return fmt.Errorf("invalid key curve for algorithm %s: %s", k.Algorithm, k.PublicKey.Curve.Params().Name)
	}
	return nil
}

// hash returns a [hash.Hash] to be used for computing the signature.
func (k *verificationKey) hash() hash.Hash {
	// TODO: Support more signature algorithms in the future.
	switch k.Algorithm {
	case "ES256":
		return sha256.New()
	default:
		panic(fmt.Sprintf("service/chat: unknown algorithm: %q", k.Algorithm))
	}
}
