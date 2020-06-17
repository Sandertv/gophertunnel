package protocol

import (
	"bytes"
)

// CreativeItem represents a creative item present in the creative inventory.
type CreativeItem struct {
	// CreativeItemNetworkID is a unique ID for the creative item. It has to be unique for each creative item
	// sent to the client. An incrementing ID per creative item does the job.
	CreativeItemNetworkID uint32
	// Item is the item that should be added to the creative inventory.
	Item ItemStack
}

// WriteCreativeEntry writes a CreativeItem x to the Buffer dst.
func WriteCreativeEntry(dst *bytes.Buffer, x CreativeItem) error {
	return chainErr(
		WriteVaruint32(dst, x.CreativeItemNetworkID),
		WriteItem(dst, x.Item),
	)
}

// CreativeEntry reads a CreativeItem x from the Buffer src.
func CreativeEntry(src *bytes.Buffer, x *CreativeItem) error {
	return chainErr(
		Varuint32(src, &x.CreativeItemNetworkID),
		Item(src, &x.Item),
	)
}
