package protocol

const (
	ClockPayloadTypeSyncState = iota
	ClockPayloadTypeInitializeRegistry
	ClockPayloadTypeAddTimeMarker
	ClockPayloadTypeRemoveTimeMarker
)

// SyncWorldClockStateData represents the state data for synchronising a world clock.
type SyncWorldClockStateData struct {
	// ClockID is the unique identifier for the clock.
	ClockID uint64
	// Time is the current time of the clock.
	Time int32
	// Paused indicates if the clock is paused.
	Paused bool
}

// Marshal encodes/decodes a SyncWorldClockStateData.
func (x *SyncWorldClockStateData) Marshal(r IO) {
	r.Varuint64(&x.ClockID)
	r.Varint32(&x.Time)
	r.Bool(&x.Paused)
}

// TimeMarkerData represents a time marker within a world clock.
type TimeMarkerData struct {
	// ID is the unique identifier for the time marker.
	ID uint64
	// Name is the name of the time marker.
	Name string
	// Time is the time at which the marker is set.
	Time int32
	// Period is the period for the time marker.
	Period int32
}

// Marshal encodes/decodes a TimeMarkerData.
func (x *TimeMarkerData) Marshal(r IO) {
	r.Varuint64(&x.ID)
	r.String(&x.Name)
	r.Varint32(&x.Time)
	r.Int32(&x.Period)
}

// WorldClockData represents a complete world clock with its time markers.
type WorldClockData struct {
	// ID is the unique identifier for the clock.
	ID uint64
	// Name is the name of the clock.
	Name string
	// Time is the current time of the clock.
	Time int32
	// Paused indicates if the clock is paused.
	Paused bool
	// TimeMarkers is a list of time markers for this clock.
	TimeMarkers []TimeMarkerData
}

// Marshal encodes/decodes a WorldClockData.
func (x *WorldClockData) Marshal(r IO) {
	r.Varuint64(&x.ID)
	r.String(&x.Name)
	r.Varint32(&x.Time)
	r.Bool(&x.Paused)
	Slice(r, &x.TimeMarkers)
}
