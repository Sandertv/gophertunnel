package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ShowStoreOffer is sent by the server to show a Marketplace store offer to a player. It opens a window
// client-side that displays the item.
// The ShowStoreOffer packet only works on the partnered servers: Servers that are not partnered will not have
// a store buttons show up in the in-game pause menu and will, as a result, not be able to open store offers
// on the client side. Sending the packet does therefore not work when using a proxy that is not connected to
// with the domain of one of the partnered servers.
type ShowStoreOffer struct {
	// OfferID is a string that identifies the offer for which a window should be opened. While typically a
	// UUID, the ID could be anything.
	OfferID string
	// ShowAll specifies if all other offers of the same 'author' as the one of the offer associated with the
	// OfferID should also be displayed, alongside the target offer.
	ShowAll bool
}

// ID ...
func (*ShowStoreOffer) ID() uint32 {
	return IDShowStoreOffer
}

// Marshal ...
func (pk *ShowStoreOffer) Marshal(w *protocol.Writer) {
	w.String(&pk.OfferID)
	w.Bool(&pk.ShowAll)
}

// Unmarshal ...
func (pk *ShowStoreOffer) Unmarshal(r *protocol.Reader) {
	r.String(&pk.OfferID)
	r.Bool(&pk.ShowAll)
}
