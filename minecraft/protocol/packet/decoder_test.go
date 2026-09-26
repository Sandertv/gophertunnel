package packet

import (
	"bytes"
	"testing"
)

func TestDecoderPayloadLimit(t *testing.T) {
	payload := []byte{3, 1, 2, 3}
	for _, algorithm := range []Compression{nil, NopCompression, FlateCompression, SnappyCompression} {
		for _, limit := range []int{2, len(payload), 0, -1} {
			for _, callback := range []bool{false, true} {
				batch := append([]byte{header}, payload...)
				if algorithm != nil {
					compressed, err := algorithm.Compress(payload)
					if err != nil {
						t.Fatal(err)
					}
					batch = append([]byte{header, byte(algorithm.EncodeCompression())}, compressed...)
				}
				decoder := NewDecoder(bytes.NewReader(batch))
				if algorithm == nil {
					decoder.maxDecompressedLen = limit
					if limit <= 0 {
						decoder.maxDecompressedLen = DefaultMaxDecompressedLen
					}
				} else {
					decoder.EnableCompression(algorithm, limit)
				}
				var count int
				var err error
				if callback {
					err = decoder.DecodeFunc(func(data []byte) error {
						count++
						if !bytes.Equal(data, payload[1:]) {
							t.Errorf("unexpected packet: %x", data)
						}
						return nil
					})
				} else {
					var packets [][]byte
					packets, err = decoder.Decode()
					count = len(packets)
				}
				if limit == 2 {
					if err == nil || count != 0 {
						t.Errorf("algorithm %v callback %v: oversized batch dispatched %d packets, error %v", algorithm, callback, count, err)
					}
				} else if err != nil || count != 1 {
					t.Errorf("algorithm %v limit %d callback %v: dispatched %d packets, error %v", algorithm, limit, callback, count, err)
				}
			}
		}
	}
}
