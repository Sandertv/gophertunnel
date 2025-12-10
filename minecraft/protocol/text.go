package protocol

const (
	TextCategoryMessageOnly = uint8(iota)
	TextCategoryAuthoredMessage
	TextCategoryMessageWithParameters
)

var textCategories = map[uint8][]string{
	TextCategoryMessageOnly:           {"raw", "tip", "systemMessage", "textObjectWhisper", "textObjectAnnouncement", "textObject"},
	TextCategoryAuthoredMessage:       {"chat", "whisper", "announcement"},
	TextCategoryMessageWithParameters: {"translate", "popup", "jukeboxPopup"},
}
