package protocol

const (
	TextCategoryMessageOnly = iota
	TextCategoryAuthoredMessage
	TextCategoryMessageWithParameters
)

var textCategories = map[int][]string{
	TextCategoryMessageOnly:           {"raw", "tip", "systemMessage", "textObjectWhisper", "textObjectAnnouncement", "textObject"},
	TextCategoryAuthoredMessage:       {"chat", "whisper", "announcement"},
	TextCategoryMessageWithParameters: {"translate", "popup", "jukeboxPopup"},
}
