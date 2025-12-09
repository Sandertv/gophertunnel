package protocol

const (
	TextCategoryMessageOnly = uint8(iota)
	TextCategoryAuthorizedMessage
	TextCategoryMessageWithParameters
)

var textCategories = map[uint8][]string{
	TextCategoryMessageOnly:           {"raw", "tip", "systemMessage", "textObjectWhisper", "textObjectAnnouncement", "textObject"},
	TextCategoryAuthorizedMessage:     {"chat", "whisper", "announcement"},
	TextCategoryMessageWithParameters: {"translation", "popup", "jukeboxPopup"},
}
