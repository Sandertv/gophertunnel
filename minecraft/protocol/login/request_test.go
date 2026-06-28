package login

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

func TestEncodeRequestTerminatesAuthJSON(t *testing.T) {
	req := &request{
		AuthenticationType: 2,
		Token:              "auth-token",
		RawToken:           "raw-token",
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
	if parsed.Token != req.Token {
		t.Fatalf("expected parsed token %q, got %q", req.Token, parsed.Token)
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

func TestEncodeTokenUsesModernTokenOnlyAuthPayload(t *testing.T) {
	key := testKey(t)
	token := testMultiplayerToken(t, key, tokenClaims{
		Claims:          jwt.Claims{Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))},
		ClientPublicKey: MarshalPublicKey(&key.PublicKey),
		XUID:            "2533274790395900",
		DisplayName:     "Steve",
		PlayFabID:       "playfab-master-id",
		PlayFabTitleID:  "20CA2",
	})

	encoded := EncodeToken(ClientData{
		ThirdPartyName: "Steve",
		ServerAddress:  "example.com:19132",
	}, key, token)
	authJSON := readAuthJSON(t, encoded)

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(authJSON)), &fields); err != nil {
		t.Fatalf("decode auth JSON: %v", err)
	}
	if _, ok := fields["Certificate"]; ok {
		t.Fatalf("expected modern token-only payload to omit Certificate, got %s", authJSON)
	}
	if _, ok := fields["chain"]; ok {
		t.Fatalf("expected modern token-only payload to omit legacy chain, got %s", authJSON)
	}
	if _, ok := fields["Token"]; !ok {
		t.Fatalf("expected modern token-only payload to include Token, got %s", authJSON)
	}
	if _, ok := fields["AuthenticationType"]; !ok {
		t.Fatalf("expected modern token-only payload to include AuthenticationType, got %s", authJSON)
	}

	parsed, err := parseLoginRequest(encoded)
	if err != nil {
		t.Fatalf("parse encoded token request: %v", err)
	}
	if parsed.Token != token {
		t.Fatalf("expected token to round trip")
	}
	if len(parsed.Certificate.Chain) != 0 {
		t.Fatalf("expected no parsed certificate chain, got %#v", parsed.Certificate.Chain)
	}
	if parsed.RawToken == "" {
		t.Fatalf("expected signed client data token")
	}
}

func TestParseTokenIdentityData(t *testing.T) {
	key := testKey(t)
	token := testMultiplayerToken(t, key, tokenClaims{
		Claims:          jwt.Claims{Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))},
		ClientPublicKey: MarshalPublicKey(&key.PublicKey),
		XUID:            "2533274790395900",
		DisplayName:     "Steve",
		PlayFabID:       "playfab-master-id",
		PlayFabTitleID:  "20CA2",
	})

	identityData, err := ParseTokenIdentityData(context.Background(), token, nil)
	if err != nil {
		t.Fatalf("parse token identity data: %v", err)
	}
	if identityData.XUID != "2533274790395900" {
		t.Fatalf("expected XUID from token, got %q", identityData.XUID)
	}
	if identityData.DisplayName != "Steve" {
		t.Fatalf("expected display name from token, got %q", identityData.DisplayName)
	}
	if identityData.PlayFabID != "playfab-master-id" {
		t.Fatalf("expected PlayFab ID from token, got %q", identityData.PlayFabID)
	}
	if identityData.PlayFabTitleID != "20CA2" {
		t.Fatalf("expected PlayFab title ID from token, got %q", identityData.PlayFabTitleID)
	}
	if identityData.Identity == "" {
		t.Fatalf("expected derived identity from XUID")
	}
}

func testKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P384(), cryptorand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func testMultiplayerToken(t *testing.T, key *ecdsa.PrivateKey, claims tokenClaims) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Key: key, Algorithm: jose.ES384}, nil)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	token, err := signJSONWebToken(signer, claims)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func readAuthJSON(t *testing.T, encoded []byte) string {
	t.Helper()
	buf := bytes.NewBuffer(encoded)
	var chainLength int32
	if err := binary.Read(buf, binary.LittleEndian, &chainLength); err != nil {
		t.Fatalf("read auth JSON length: %v", err)
	}
	return string(buf.Next(int(chainLength)))
}
