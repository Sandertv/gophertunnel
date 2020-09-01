package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ServerSettingsResponse is optionally sent by the server in response to a ServerSettingsRequest from the
// client. It is structured the same as a ModalFormRequest packet, and if filled out correctly, will show
// a specific tab for the server in the settings of the client. A ModalFormResponse packet is sent by the
// client in response to a ServerSettingsResponse, when the client fills out the settings and closes the
// settings again.
type ServerSettingsResponse struct {
	// FormID is an ID used to identify the form. The ID is saved by the client and sent back when the player
	// submits the form, so that the server can identify which form was submitted.
	FormID uint32
	// FormData is a JSON encoded object of form data. The content of the object differs, depending on the
	// type of the form sent, which is also set in the JSON.
	FormData []byte
}

// ID ...
func (*ServerSettingsResponse) ID() uint32 {
	return IDServerSettingsResponse
}

// Marshal ...
func (pk *ServerSettingsResponse) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.FormID)
	w.ByteSlice(&pk.FormData)
}

// Unmarshal ...
func (pk *ServerSettingsResponse) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.FormID)
	r.ByteSlice(&pk.FormData)
}
