package packet

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
)

// LevelEventGeneric is sent by the server to send a 'generic' level event to the client. This packet sends an
// NBT serialised object and may for that reason be used for any event holding additional data.
type LevelEventGeneric struct {
	// EventID is a unique identifier that identifies the event called. The data that follows has fields in
	// the NBT depending on what event it is.
	EventID int32
	// SerialisedEventData is a network little endian serialised object of event data, with fields that vary
	// depending on EventID.
	SerialisedEventData []byte
}

// ID ...
func (pk *LevelEventGeneric) ID() uint32 {
	return IDLevelEventGeneric
}

// Marshal ...
func (pk *LevelEventGeneric) Marshal(buf *bytes.Buffer) {
	_ = protocol.WriteVarint32(buf, pk.EventID)
	_, _ = buf.Write(pk.SerialisedEventData)
}

// Unmarshal ...
func (pk *LevelEventGeneric) Unmarshal(buf *bytes.Buffer) error {
	if err := protocol.Varint32(buf, &pk.EventID); err != nil {
		return err
	}
	pk.SerialisedEventData = buf.Next(math.MaxInt32)
	return nil
}
