package discovery

import (
	"encoding/hex"
	"fmt"
	"io"
)

type ResponsePacket struct {
	ApplicationData []byte
}

func (*ResponsePacket) ID() uint16 { return IDResponsePacket }

func (pk *ResponsePacket) Read(r io.Reader) error {
	data, err := readBytes[uint32](r)
	if err != nil {
		return fmt.Errorf("read application data: %w", err)
	}
	n, err := hex.Decode(data, data)
	if err != nil {
		return fmt.Errorf("decode application data: %w", err)
	}
	pk.ApplicationData = data[:n]
	return nil
}

func (pk *ResponsePacket) Write(w io.Writer) {
	data := make([]byte, hex.EncodedLen(len(pk.ApplicationData)))
	hex.Encode(data, pk.ApplicationData)
	writeBytes[uint32](w, data)
}
