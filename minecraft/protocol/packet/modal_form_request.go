package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ModalFormRequest is sent by the server to make the client open a form. This form may be either a modal form
// which has two options, a menu form for a selection of options and a custom form for properties.
type ModalFormRequest struct {
	// FormID is an ID used to identify the form. The ID is saved by the client and sent back when the player
	// submits the form, so that the server can identify which form was submitted.
	FormID uint32
	// FormData is a JSON encoded object of form data. The content of the object differs, depending on the
	// type of the form sent, which is also set in the JSON.
	FormData []byte
}

// ID ...
func (*ModalFormRequest) ID() uint32 {
	return IDModalFormRequest
}

// Marshal ...
func (pk *ModalFormRequest) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.FormID)
	w.ByteSlice(&pk.FormData)
}

// Unmarshal ...
func (pk *ModalFormRequest) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.FormID)
	r.ByteSlice(&pk.FormData)
}
