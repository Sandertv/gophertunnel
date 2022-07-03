package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// UpdateAdventureSettings is a packet sent from the server to the client to update the adventure settings of the player.
// It, along with the UpdateAbilities packet, are replacements of the AdventureSettings packet since v1.19.10.
type UpdateAdventureSettings struct {
	// NoPvM is a boolean indicating whether the player is allowed to fight mobs or not.
	NoPvM bool
	// NoMvP is a boolean indicating whether mobs are allowed to fight the player or not. It is unclear why this is sent
	// to the client.
	NoMvP bool
	// ImmutableWorld is a boolean indicating whether the player is allowed to modify the world or not.
	ImmutableWorld bool
	// ShowNameTags is a boolean indicating whether player name tags are shown or not.
	ShowNameTags bool
	// AutoJump is a boolean indicating whether the player is allowed to jump automatically or not.
	AutoJump bool
}

// ID ...
func (*UpdateAdventureSettings) ID() uint32 {
	return IDUpdateAdventureSettings
}

// Marshal ...
func (pk *UpdateAdventureSettings) Marshal(w *protocol.Writer) {
	w.Bool(&pk.NoPvM)
	w.Bool(&pk.NoMvP)
	w.Bool(&pk.ImmutableWorld)
	w.Bool(&pk.ShowNameTags)
	w.Bool(&pk.AutoJump)
}

// Unmarshal ...
func (pk *UpdateAdventureSettings) Unmarshal(r *protocol.Reader) {
	r.Bool(&pk.NoPvM)
	r.Bool(&pk.NoMvP)
	r.Bool(&pk.ImmutableWorld)
	r.Bool(&pk.ShowNameTags)
	r.Bool(&pk.AutoJump)
}
