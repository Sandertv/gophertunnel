package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	BookActionReplacePage = iota
	BookActionAddPage
	BookActionDeletePage
	BookActionSwapPages
	BookActionSign
)

// BookEdit is sent by the client when it edits a book. It is sent each time a modification was made and the
// player stops its typing 'session', rather than simply after closing the book.
type BookEdit struct {
	// ActionType is the type of the book edit action. The data obtained depends on what type this is. The
	// action type is one of the constants above.
	ActionType byte
	// InventorySlot is the slot in which the book that was edited may be found. Typically, the server should
	// check if this slot matches the held item slot of the player.
	InventorySlot byte
	// PageNumber is the number of the page that the book edit action concerns. It applies for all actions
	// but the BookActionSign. In BookActionSwapPages, it is one of the pages that was swapped.
	PageNumber byte
	// SecondaryPageNumber is the page number of the second page that the action concerned. It is only set for
	// the BookActionSwapPages action, in which case it is the other page that is swapped.
	SecondaryPageNumber byte
	// Text is the text that was written in a particular page of the book. It applies for the
	// BookActionAddPage and BookActionReplacePage only.
	Text string
	// PhotoName is the name of the photo on the page in the book. It applies for the BookActionAddPage and
	// BookActionReplacePage only.
	// Unfortunately, the functionality of this field was removed from the default Minecraft Bedrock Edition.
	// It is still available on Education Edition.
	PhotoName string
	// Title is the title that the player has given the book. It applies only for the BookActionSign action.
	Title string
	// Author is the author that the player has given the book. It applies only for the BookActionSign action.
	// Note that the author may be freely changed, so no assumptions can be made on if the author is actually
	// the name of a player.
	Author string
	// XUID is the XBOX Live User ID of the player that edited the book. The field is rather pointless, as the
	// server is already aware of the XUID of the player anyway.
	XUID string
}

// ID ...
func (*BookEdit) ID() uint32 {
	return IDBookEdit
}

func (pk *BookEdit) Marshal(io protocol.IO) {
	io.Uint8(&pk.ActionType)
	io.Uint8(&pk.InventorySlot)
	switch pk.ActionType {
	case BookActionReplacePage, BookActionAddPage:
		io.Uint8(&pk.PageNumber)
		io.String(&pk.Text)
		io.String(&pk.PhotoName)
	case BookActionDeletePage:
		io.Uint8(&pk.PageNumber)
	case BookActionSwapPages:
		io.Uint8(&pk.PageNumber)
		io.Uint8(&pk.SecondaryPageNumber)
	case BookActionSign:
		io.String(&pk.Title)
		io.String(&pk.Author)
		io.String(&pk.XUID)
	default:
		io.UnknownEnumOption(pk.ActionType, "book edit action type")
	}
}
