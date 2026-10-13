package protocol

// SignedChat is a chat message signed by the Chat Service. It must be relayed unmodified by the server.
type SignedChat struct {
	// KeyID identifies which published Chat Service key signed the message. It is empty when unsigned.
	KeyID string
	// SenderID is the identity the Chat Service signed against. It is empty when unsigned.
	SenderID string
	// RoomID is the room the signature was minted for. It is empty when unsigned.
	RoomID string
	// MessageID is the client generated ID of the message. It is empty when unsigned.
	MessageID string
	// ScanText is the message the Chat Service scanned and signed.
	ScanText string
	// RedactedText is the redacted form of the message. It is empty if nothing needed redacting.
	RedactedText string
	// Signature is the Chat Service signature over the fields above. It is empty when unsigned.
	Signature string
}

// Marshal encodes/decodes a SignedChat.
func (x *SignedChat) Marshal(r IO) {
	r.String(&x.KeyID)
	r.String(&x.SenderID)
	r.String(&x.RoomID)
	r.String(&x.MessageID)
	r.String(&x.ScanText)
	r.String(&x.RedactedText)
	r.String(&x.Signature)
}

// ChatFormatting holds optional server decoration rendered around the signed name and message of a chat
// message.
type ChatFormatting struct {
	// NamePrefix is rendered before the author name.
	NamePrefix Optional[string]
	// NamePostfix is rendered after the author name.
	NamePostfix Optional[string]
	// MessagePrefix is rendered before the message.
	MessagePrefix Optional[string]
	// MessagePostfix is rendered after the message.
	MessagePostfix Optional[string]
}

// Marshal encodes/decodes a ChatFormatting.
func (x *ChatFormatting) Marshal(r IO) {
	OptionalFunc(r, &x.NamePrefix, r.String)
	OptionalFunc(r, &x.NamePostfix, r.String)
	OptionalFunc(r, &x.MessagePrefix, r.String)
	OptionalFunc(r, &x.MessagePostfix, r.String)
}
