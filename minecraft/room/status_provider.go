package room

type StatusProvider interface {
	RoomStatus() Status
}
