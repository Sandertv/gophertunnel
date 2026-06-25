package packet

import (
	"bytes"
	"testing"
)

func TestDecoderSkipsEmptyPacketsWhenBatchPacketLimitDisabled(t *testing.T) {
	payload := []byte{42}
	batch := []byte{header, 0, byte(len(payload))}
	batch = append(batch, payload...)

	strictDecoder := NewDecoder(bytes.NewReader(batch))
	if _, err := strictDecoder.Decode(); err == nil {
		t.Fatal("Decode strict = nil, want empty packet error")
	}

	serverDecoder := NewDecoder(bytes.NewReader(batch))
	serverDecoder.DisableBatchPacketLimit()
	packets, err := serverDecoder.Decode()
	if err != nil {
		t.Fatalf("Decode server: %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("len(packets) = %d, want 1", len(packets))
	}
	if !bytes.Equal(packets[0], payload) {
		t.Fatalf("packets[0] = %v, want %v", packets[0], payload)
	}
}
