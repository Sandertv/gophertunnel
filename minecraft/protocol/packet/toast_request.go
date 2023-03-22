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

// Marshal ...
func (pk *ToastRequest) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *ToastRequest) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *ToastRequest) marshal(r protocol.IO) {
	r.String(&pk.Title)
	r.String(&pk.Message)
}
