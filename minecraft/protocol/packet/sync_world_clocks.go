package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// SyncWorldClocks is sent by the server to initialise and synchronise world clocks with the client.
type SyncWorldClocks struct {
	// PayloadType is the type of clock payload. It is one of the protocol.ClockPayloadType constants.
	PayloadType uint32
	// SyncStates is set if PayloadType is ClockPayloadTypeSyncState.
	SyncStates []protocol.SyncWorldClockStateData
	// Clocks is set if PayloadType is ClockPayloadTypeInitializeRegistry.
	Clocks []protocol.WorldClockData
	// AddClockID is the clock ID for adding time markers, set if PayloadType is ClockPayloadTypeAddTimeMarker.
	AddClockID uint64
	// AddTimeMarkers is the list of time markers to add, set if PayloadType is ClockPayloadTypeAddTimeMarker.
	AddTimeMarkers []protocol.TimeMarkerData
	// RemoveClockID is the clock ID for removing time markers, set if PayloadType is ClockPayloadTypeRemoveTimeMarker.
	RemoveClockID uint64
	// RemoveTimeMarkerIDs is the list of time marker IDs to remove.
	RemoveTimeMarkerIDs []uint64
}

// ID ...
func (*SyncWorldClocks) ID() uint32 {
	return IDSyncWorldClocks
}

func (pk *SyncWorldClocks) Marshal(io protocol.IO) {
	io.Varuint32(&pk.PayloadType)
	switch pk.PayloadType {
	case protocol.ClockPayloadTypeSyncState:
		protocol.Slice(io, &pk.SyncStates)
	case protocol.ClockPayloadTypeInitializeRegistry:
		protocol.Slice(io, &pk.Clocks)
	case protocol.ClockPayloadTypeAddTimeMarker:
		io.Varuint64(&pk.AddClockID)
		protocol.Slice(io, &pk.AddTimeMarkers)
	case protocol.ClockPayloadTypeRemoveTimeMarker:
		io.Varuint64(&pk.RemoveClockID)
		protocol.FuncSlice(io, &pk.RemoveTimeMarkerIDs, io.Varuint64)
	default:
		io.UnknownEnumOption(pk.PayloadType, "clock payload type")
	}
}
