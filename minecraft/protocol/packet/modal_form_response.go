package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	ModalFormCancelReasonUserClosed = iota
	ModalFormCancelReasonUserBusy
)

// ModalFormResponse is sent by the client in response to a ModalFormRequest, after the player has submitted
// the form sent. It contains the options/properties selected by the player, or a JSON encoded 'null' if
// the form was closed by clicking the X at the top right corner of the form.
type ModalFormResponse struct {
	// FormID is the form ID of the form the client has responded to. It is the same as the ID sent in the
	// ModalFormRequest, and may be used to identify which form was submitted.
	FormID uint32
	// HasResponseData is true if the client provided response data.
	HasResponseData bool
	// ResponseData is a JSON encoded value representing the response of the player. For a modal form, the response is
	// either true or false, for a menu form, the response is an integer specifying the index of the button clicked, and
	// for a custom form, the response is an array containing a value for each element.
	ResponseData []byte
	// HasCancelReason is true if the client provided a reason for the form being cancelled.
	HasCancelReason bool
	// CancelReason represents the reason why the form was cancelled. It is one of the constants above.
	CancelReason uint8
}

// ID ...
func (*ModalFormResponse) ID() uint32 {
	return IDModalFormResponse
}

// Marshal ...
func (pk *ModalFormResponse) Marshal(w *protocol.Writer) {
	w.Varuint32(&pk.FormID)
	w.Bool(&pk.HasResponseData)
	if pk.HasResponseData {
		w.ByteSlice(&pk.ResponseData)
	}
	w.Bool(&pk.HasCancelReason)
	if pk.HasCancelReason {
		w.Uint8(&pk.CancelReason)
	}
}

// Unmarshal ...
func (pk *ModalFormResponse) Unmarshal(r *protocol.Reader) {
	r.Varuint32(&pk.FormID)
	r.Bool(&pk.HasResponseData)
	if pk.HasResponseData {
		r.ByteSlice(&pk.ResponseData)
	}
	r.Bool(&pk.HasCancelReason)
	if pk.HasCancelReason {
		r.Uint8(&pk.CancelReason)
	}
}
