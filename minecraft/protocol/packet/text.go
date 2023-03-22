package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	TextTypeRaw = iota
	TextTypeChat
	TextTypeTranslation
	TextTypePopup
	TextTypeJukeboxPopup
	TextTypeTip
	TextTypeSystem
	TextTypeWhisper
	TextTypeAnnouncement
	TextTypeObjectWhisper
	TextTypeObject
	TextTypeObjectAnnouncement
)

// Text is sent by the client to the server to send chat messages, and by the server to the client to forward
// or send messages, which may be chat, popups, tips etc.
type Text struct {
	// TextType is the type of the text sent. When a client sends this to the server, it should always be
	// TextTypeChat. If the server sends it, it may be one of the other text types above.
	TextType byte
	// NeedsTranslation specifies if any of the messages need to be translated. It seems that where % is found
	// in translatable text types, these are translated regardless of this bool. Translatable text types
	// include TextTypeTranslation, TextTypeTip, TextTypePopup and TextTypeJukeboxPopup.
	NeedsTranslation bool
	// SourceName is the name of the source of the messages. This source is displayed in text types such as
	// the TextTypeChat and TextTypeWhisper, where typically the username is shown.
	SourceName string
	// Message is the message of the packet. This field is set for each TextType and is the main component of
	// the packet.
	Message string
	// Parameters is a list of parameters that should be filled into the message. These parameters are only
	// written if the type of the packet is TextTypeTranslation, TextTypeTip, TextTypePopup or TextTypeJukeboxPopup.
	Parameters []string
	// XUID is the XBOX Live user ID of the player that sent the message. It is only set for packets of
	// TextTypeChat. When sent to a player, the player will only be shown the chat message if a player with
	// this XUID is present in the player list and not muted, or if the XUID is empty.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
}

// ID ...
func (*Text) ID() uint32 {
	return IDText
}

// Marshal ...
func (pk *Text) Marshal(w *protocol.Writer) {
	pk.marshal(w)
}

// Unmarshal ...
func (pk *Text) Unmarshal(r *protocol.Reader) {
	pk.marshal(r)
}

func (pk *Text) marshal(r protocol.IO) {
	r.Uint8(&pk.TextType)
	r.Bool(&pk.NeedsTranslation)
	switch pk.TextType {
	case TextTypeChat, TextTypeWhisper, TextTypeAnnouncement:
		r.String(&pk.SourceName)
		r.String(&pk.Message)
	case TextTypeRaw, TextTypeTip, TextTypeSystem, TextTypeObject, TextTypeObjectWhisper, TextTypeObjectAnnouncement:
		r.String(&pk.Message)
	case TextTypeTranslation, TextTypePopup, TextTypeJukeboxPopup:
		r.String(&pk.Message)
		protocol.FuncSlice(r, &pk.Parameters, r.String)
	}
	r.String(&pk.XUID)
	r.String(&pk.PlatformChatID)
}
