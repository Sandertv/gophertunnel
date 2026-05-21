package login

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestEncodeRequestTerminatesChainJSON(t *testing.T) {
	req := &request{
		Certificate: certificate{Chain: chain{"chain-entry"}},
		RawToken:    "raw-token",
		Legacy:      true,
	}
	encoded := encodeRequest(req)
	buf := bytes.NewBuffer(encoded)

	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		t.Fatalf("read chain length: %v", err)
	}
	chainData := buf.Next(int(chainLength))
	if !bytes.HasSuffix(chainData, []byte{'\n'}) {
		t.Fatalf("expected chain JSON to end with newline, got %q", chainData)
	}
	if !json.Valid(chainData) {
		t.Fatalf("expected chain JSON to remain valid, got %q", chainData)
	}

	var rawLength int32
	if err := binary.Read(buf, binary.LittleEndian, &rawLength); err != nil {
		t.Fatalf("read raw token length: %v", err)
	}
	rawToken := buf.Next(int(rawLength))
	if string(rawToken) != req.RawToken {
		t.Fatalf("expected raw token %q, got %q", req.RawToken, rawToken)
	}
	if bytes.HasSuffix(rawToken, []byte{'\n'}) {
		t.Fatalf("expected raw token to remain unterminated, got %q", rawToken)
	}
	if buf.Len() != 0 {
		t.Fatalf("unexpected trailing bytes: %x", buf.Bytes())
	}

	parsed, err := parseLoginRequest(encoded)
	if err != nil {
		t.Fatalf("parse encoded request: %v", err)
	}
	if len(parsed.Certificate.Chain) != 1 || parsed.Certificate.Chain[0] != "chain-entry" {
		t.Fatalf("unexpected chain: %#v", parsed.Certificate.Chain)
	}
	if parsed.RawToken != req.RawToken {
		t.Fatalf("expected parsed raw token %q, got %q", req.RawToken, parsed.RawToken)
	}
}

func TestSignJSONWebTokenTerminatesPayloadJSON(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), cryptorand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Key: key, Algorithm: jose.ES384}, nil)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	token, err := signJSONWebToken(signer, map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected compact JWT with 3 parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !bytes.HasSuffix(payload, []byte{'\n'}) {
		t.Fatalf("expected signed payload JSON to end with newline, got %q", payload)
	}
	if !json.Valid(payload) {
		t.Fatalf("expected payload JSON to remain valid, got %q", payload)
	}
}
