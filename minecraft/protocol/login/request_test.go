package login

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
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
