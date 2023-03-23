package packet

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// ToastRequest is a packet sent from the server to the client to display a toast to the top of the screen. These toasts
// are the same as the ones seen when, for example, loading a new resource pack or obtaining an achievement.
type ToastRequest struct {
	// Title is the title of the toast.
	Title string
	// Message is the message that the toast may contain alongside the title.
	Message string
}

// ID ...
func (*ToastRequest) ID() uint32 {
	return IDToastRequest
}

func (pk *ToastRequest) Marshal(io protocol.IO) {
	io.String(&pk.Title)
	io.String(&pk.Message)
}
