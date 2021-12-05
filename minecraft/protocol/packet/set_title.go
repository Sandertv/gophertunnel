package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	TitleActionClear = iota
	TitleActionReset
	TitleActionSetTitle
	TitleActionSetSubtitle
	TitleActionSetActionBar
	TitleActionSetDurations
	TitleActionTitleTextObject
	TitleActionSubtitleTextObject
	TitleActionActionbarTextObject
)

// SetTitle is sent by the server to make a title, subtitle or action bar shown to a player. It has several
// fields that allow setting the duration of the titles.
type SetTitle struct {
	// ActionType is the type of the action that should be executed upon the title of a player. It is one of
	// the constants above and specifies the response of the client to the packet.
	ActionType int32
	// Text is the text of the title, which has a different meaning depending on the ActionType that the
	// packet has. The text is the text of a title, subtitle or action bar, depending on the type set.
	Text string
	// FadeInDuration is the duration that the title takes to fade in on the screen of the player. It is
	// measured in 20ths of a second (AKA in ticks).
	FadeInDuration int32
	// RemainDuration is the duration that the title remains on the screen of the player. It is measured in
	// 20ths of a second (AKA in ticks).
	RemainDuration int32
	// FadeOutDuration is the duration that the title takes to fade out of the screen of the player. It is
	// measured in 20ths of a second (AKA in ticks).
	FadeOutDuration int32
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account. It is empty if the user is not logged into its XBL account.
	XUID string
	// PlatformOnlineID is either a uint64 or an empty string.
	PlatformOnlineID string
}

// ID ...
func (*SetTitle) ID() uint32 {
	return IDSetTitle
}

// Marshal ...
func (pk *SetTitle) Marshal(w *protocol.Writer) {
	w.Varint32(&pk.ActionType)
	w.String(&pk.Text)
	w.Varint32(&pk.FadeInDuration)
	w.Varint32(&pk.RemainDuration)
	w.Varint32(&pk.FadeOutDuration)
	w.String(&pk.XUID)
	w.String(&pk.PlatformOnlineID)
}

// Unmarshal ...
func (pk *SetTitle) Unmarshal(r *protocol.Reader) {
	r.Varint32(&pk.ActionType)
	r.String(&pk.Text)
	r.Varint32(&pk.FadeInDuration)
	r.Varint32(&pk.RemainDuration)
	r.Varint32(&pk.FadeOutDuration)
	r.String(&pk.XUID)
	r.String(&pk.PlatformOnlineID)
}
