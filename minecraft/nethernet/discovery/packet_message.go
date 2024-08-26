package discovery

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessagePacket struct {
	RecipientID uint64
	Data        string
}

func (*MessagePacket) ID() uint16 { return IDMessagePacket }

func (pk *MessagePacket) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &pk.RecipientID); err != nil {
		return fmt.Errorf("read recipient ID: %w", err)
	}
	data, err := readBytes[uint32](r)
	if err != nil {
		return fmt.Errorf("read data: %w", err)
	}
	pk.Data = string(data)
	return nil
}

func (pk *MessagePacket) Write(w io.Writer) {
	_ = binary.Write(w, binary.LittleEndian, pk.RecipientID)
	writeBytes[uint32](w, []byte(pk.Data))
}
