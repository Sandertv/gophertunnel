package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// LevelEventGeneric is sent by the server to send a 'generic' level event to the client. This packet sends an
// NBT serialised object and may for that reason be used for any event holding additional data.
type LevelEventGeneric struct {
	// EventID is a unique identifier that identifies the event called. The data that follows has fields in
	// the NBT depending on what event it is.
	EventID int32
	// SerialisedEventData is a network little endian serialised object of event data, with fields that vary
	// depending on EventID.
	// Unlike many other NBT structures, this data is not actually in a compound but just loosely floating
	// NBT tags. To decode using the nbt package, you would need to append 0x0a00 at the start (compound id
	// and name length) and add 0x00 at the end, to manually wrap it in a compound. Likewise, you would have
	// to remove these bytes when encoding.
	// Example of the resulting data with an EventID of 2026:
	// TAG_Compound({
	//    'pos15x': TAG_Float(198),
	//    'pos11x': TAG_Float(201),
	//    'pos6y': TAG_Float(65),
	//    'pos13y': TAG_Float(64),
	//    'pos17z': TAG_Float(36),
	//    'pos8y': TAG_Float(65),
	//    'originY': TAG_Float(65.06125),
	//    'pos10z': TAG_Float(37),
	//    'pos13x': TAG_Float(201),
	//    'pos7y': TAG_Float(65),
	//    'pos9x': TAG_Float(203),
	//    'pos11y': TAG_Float(64),
	//    'pos15y': TAG_Float(65),
	//    'pos15z': TAG_Float(40),
	//    'pos7z': TAG_Float(41),
	//    'pos8x': TAG_Float(198),
	//    'pos13z': TAG_Float(40),
	//    'pos1z': TAG_Float(37),
	//    'pos6z': TAG_Float(42),
	//    'size': TAG_Int(18),
	//    'pos0x': TAG_Float(204),
	//    'pos12x': TAG_Float(200),
	//    'pos2x': TAG_Float(204),
	//    'pos9z': TAG_Float(37),
	//    'pos16y': TAG_Float(64),
	//    'pos5x': TAG_Float(204),
	//    'pos5y': TAG_Float(64),
	//    'pos17x': TAG_Float(202),
	//    'pos3y': TAG_Float(64),
	//    'pos3z': TAG_Float(36),
	//    'radius': TAG_Float(4),
	//    'pos0z': TAG_Float(38),
	//    'pos4z': TAG_Float(36),
	//    'pos8z': TAG_Float(38),
	//    'pos1x': TAG_Float(204),
	//    'pos0y': TAG_Float(64),
	//    'pos14z': TAG_Float(39),
	//    'pos16z': TAG_Float(40),
	//    'pos2y': TAG_Float(63),
	//    'pos6x': TAG_Float(203),
	//    'pos10x': TAG_Float(205),
	//    'pos12y': TAG_Float(64),
	//    'pos1y': TAG_Float(64),
	//    'pos14x': TAG_Float(200),
	//    'pos3x': TAG_Float(204),
	//    'pos9y': TAG_Float(64),
	//    'pos4y': TAG_Float(63),
	//    'pos10y': TAG_Float(63),
	//    'pos12z': TAG_Float(38),
	//    'pos16x': TAG_Float(202),
	//    'originX': TAG_Float(202.48654),
	//    'pos14y': TAG_Float(62),
	//    'pos17y': TAG_Float(62),
	//    'pos5z': TAG_Float(35),
	//    'pos4x': TAG_Float(204),
	//    'pos7x': TAG_Float(203),
	//    'originZ': TAG_Float(38.297028),
	//    'pos11z': TAG_Float(38),
	//    'pos2z': TAG_Float(39),
	// })
	// The 'originX', 'originY' and 'originZ' fields are present in every event and serve as a replacement for
	// a Position field in this packet.
	SerialisedEventData []byte
}

// ID ...
func (pk *LevelEventGeneric) ID() uint32 {
	return IDLevelEventGeneric
}

// Marshal ...
func (pk *LevelEventGeneric) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.EventID)
	w.Bytes(&pk.SerialisedEventData)
}

// Unmarshal ...
func (pk *LevelEventGeneric) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.EventID)
	r.Bytes(&pk.SerialisedEventData)
}
