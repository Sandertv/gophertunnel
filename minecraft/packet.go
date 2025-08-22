package minecraft

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// packetData holds the data of a Minecraft packet.
type packetData struct {
	h       *packet.Header
	full    []byte
	payload *bytes.Buffer
}

// parseData parses the packet data slice passed into a packetData struct.
func parseData(data []byte, conn *Conn) *packetData {
	buf := bytes.NewBuffer(data)
	header := &packet.Header{}
	if err := header.Read(buf); err != nil {
		// We don't return this as an error as it's not in the hand of the user to control this. Instead,
		// we return to reading a new packet.
		return nil
	}
	if conn.packetFunc != nil {
		// The packet func was set, so we call it.
		conn.packetFunc(*header, buf.Bytes(), conn.RemoteAddr(), conn.LocalAddr())
	}
	return &packetData{h: header, full: data, payload: buf}
}

type unknownPacketError struct {
	id uint32
}

func (err unknownPacketError) Error() string {
	return fmt.Sprintf("unexpected packet (ID=%v)", err.id)
}

// decode decodes the packet payload held in the packetData and returns the packet.Packet decoded.
func (p *packetData) decode(conn *Conn) (pks []packet.Packet, err error) {
	// Attempt to fetch the packet with the right packet ID from the pool.
	pkFunc, ok := conn.pool[p.h.PacketID]
	var pk packet.Packet
	if !ok {
		// No packet with the ID. This may be a custom packet of some sorts.
		pk = &packet.Unknown{PacketID: p.h.PacketID}
		if conn.disconnectOnUnknownPacket {
			_ = conn.Close()
			return nil, unknownPacketError{id: p.h.PacketID}
		}
	} else {
		pk = pkFunc()
	}

	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			err = fmt.Errorf("decode packet %T: %w", pk, recoveredErr.(error))
		}
		if err != nil && !errors.Is(err, unknownPacketError{}) && conn.disconnectOnInvalidPacket {
			_ = conn.Close()
		}
	}()

	r := conn.proto.NewReader(p.payload, conn.shieldID.Load(), conn.readerLimits)
	pk.Marshal(r)
	if p.payload.Len() != 0 {
		err = fmt.Errorf("decode packet %T: %v unread bytes left: 0x%x", pk, p.payload.Len(), p.payload.Bytes())
	}
	if conn.disconnectOnInvalidPacket && err != nil {
		return nil, err
	}
	return conn.proto.ConvertToLatest(pk, conn), err
}
