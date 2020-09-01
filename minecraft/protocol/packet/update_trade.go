package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateTrade is sent by the server to update the trades offered by a villager to a player. It is sent at the
// moment that a player interacts with a villager.
type UpdateTrade struct {
	// WindowID is the ID that identifies the trading window that the client currently has opened.
	WindowID byte
	// WindowType is an identifier specifying the type of the window opened. In vanilla, it appears this is
	// always filled out with 15.
	WindowType byte
	// Size is the amount of trading options that the villager has.
	Size int32
	// TradeTier is the tier of the villager that the player is trading with. The tier starts at 0 with a
	// first two offers being available, after which two additional offers are unlocked each time the tier
	// becomes one higher.
	TradeTier int32
	// VillagerUniqueID is the unique ID of the villager entity that the player is trading with. The
	// TradeTier sent above applies to this villager.
	VillagerUniqueID int64
	// EntityUniqueID is the unique ID of the entity (usually a player) for which the trades are updated. The
	// updated trades may apply only to this entity.
	EntityUniqueID int64
	// DisplayName is the name displayed at the top of the trading UI. It is usually used to represent the
	// profession of the villager in the UI.
	DisplayName string
	// NewTradeUI specifies if the villager should be using the new trade UI (The one added in 1.11.) rather
	// than the old one. This should usually be set to true.
	NewTradeUI bool
	// DemandBasedPrices specifies if the prices of the villager's offers are modified by an increase in
	// demand for the item. (A mechanic added in 1.11.) Buying more of the same item will increase the price
	// of that particular item.
	DemandBasedPrices bool
	// SerialisedOffers is a network NBT serialised compound of offers that the villager has.
	SerialisedOffers []byte
}

// ID ...
func (*UpdateTrade) ID() uint32 {
	return IDUpdateTrade
}

// Marshal ...
func (pk *UpdateTrade) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.WindowID)
	w.Uint8(&pk.WindowType)
	w.Varint32(&pk.Size)
	w.Varint32(&pk.TradeTier)
	w.Varint64(&pk.VillagerUniqueID)
	w.Varint64(&pk.EntityUniqueID)
	w.String(&pk.DisplayName)
	w.Bool(&pk.NewTradeUI)
	w.Bool(&pk.DemandBasedPrices)
	w.Bytes(&pk.SerialisedOffers)
}

// Unmarshal ...
func (pk *UpdateTrade) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.WindowID)
	r.Uint8(&pk.WindowType)
	r.Varint32(&pk.Size)
	r.Varint32(&pk.TradeTier)
	r.Varint64(&pk.VillagerUniqueID)
	r.Varint64(&pk.EntityUniqueID)
	r.String(&pk.DisplayName)
	r.Bool(&pk.NewTradeUI)
	r.Bool(&pk.DemandBasedPrices)
	r.Bytes(&pk.SerialisedOffers)
}
