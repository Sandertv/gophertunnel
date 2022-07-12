package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PositionTrackingDBBroadcastActionUpdate = iota
	PositionTrackingDBBroadcastActionDestroy
	PositionTrackingDBBroadcastActionNotFound
)

// PositionTrackingDBServerBroadcast is sent by the server in response to the
// PositionTrackingDBClientRequest packet. This packet is, as of 1.16, currently only used for lodestones. The
// server maintains a database with tracking IDs and their position and dimension. The client will request
// these tracking IDs, (NBT tag set on the lodestone compass with the tracking ID?) and the server will
// respond with the status of those tracking IDs.
// What is actually done with the data sent depends on what the client chooses to do with it. For the
// lodestone compass, it is used to make the compass point towards lodestones and to make it spin if the
// lodestone at a position is no longer there.
type PositionTrackingDBServerBroadcast struct {
	// BroadcastAction specifies the status of the position tracking DB response. It is one of the constants
	// above, specifying the result of the request with the ID below.
	// The Update action is sent for setting the position of a lodestone compass, the Destroy and NotFound to
	// indicate that there is not (no longer) a lodestone at that position.
	BroadcastAction byte
	// TrackingID is the ID of the PositionTrackingDBClientRequest packet that this packet was in response to.
	// The tracking ID is also present as the 'id' field in the SerialisedData field.
	TrackingID int32
	// Payload is a network little endian compound tag holding the data retrieved from the position tracking DB.
	// An example data structure sent if BroadcastAction is of the type Update:
	// TAG_Compound({
	//        'version': TAG_Byte(0x01),
	//        'dim': TAG_Int(0),
	//        'id': TAG_String(0x00000001),
	//        'pos': TAG_List<TAG_Int>({
	//                -299,
	//                86,
	//                74,
	//        }),
	//        'status': TAG_Byte(0x00), // 0x00 for updating, 0x02 for not found/block destroyed.
	// })
	Payload map[string]any
}

// ID ...
func (*PositionTrackingDBServerBroadcast) ID() uint32 {
	return IDPositionTrackingDBServerBroadcast
}

// Marshal ...
func (pk *PositionTrackingDBServerBroadcast) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.BroadcastAction)
	w.Varint32(&pk.TrackingID)
	w.NBT(&pk.Payload, nbt.NetworkLittleEndian)
}

// Unmarshal ...
func (pk *PositionTrackingDBServerBroadcast) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.BroadcastAction)
	r.Varint32(&pk.TrackingID)
	r.NBT(&pk.Payload, nbt.NetworkLittleEndian)
}
