package protocol

import (
	"bytes"
	"github.com/google/uuid"
)

// PlayerListEntry is an entry found in the PlayerList packet. It represents a single player using the UUID
// found in the entry, and contains several properties such as the skin.
type PlayerListEntry struct {
	// UUID is the UUID of the player as sent in the Login packet when the client joined the server. It must
	// match this UUID exactly for the correct XBOX Live icon to show up in the list.
	UUID uuid.UUID
	// EntityUniqueID is the unique entity ID of the player. This ID typically stays consistent during the
	// lifetime of a world, but servers often send the runtime ID for this.
	EntityUniqueID int64
	// Username is the username that is shown in the player list of the player that obtains a PlayerList
	// packet with this entry. It does not have to be the same as the actual username of the player.
	Username string
	// SkinID is a unique ID produced for the skin, for example 'c18e65aa-7b21-4637-9b63-8ad63622ef01_Alex'
	// for the default Alex skin.
	SkinID string
	// SkinData is a byte slice of 64*32*4, 64*64*4 or 128*128*4 bytes. It is a RGBA ordered byte
	// representation of the skin colours.
	SkinData []byte
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData []byte
	// SkinGeometryName is the geometry name of the skin geometry above. This name must be equal to one of the
	// outer names found in the SkinGeometry, so that the client can find the correct geometry data.
	SkinGeometryName string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry []byte
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
}

// WritePlayerAddEntry writes a PlayerListEntry x to Buffer buf in a way that adds the player to the list.
func WritePlayerAddEntry(buf *bytes.Buffer, x PlayerListEntry) error {
	return chainErr(
		WriteUUID(buf, x.UUID),
		WriteVarint64(buf, x.EntityUniqueID),
		WriteString(buf, x.Username),
		WriteString(buf, x.SkinID),
		WriteByteSlice(buf, x.SkinData),
		WriteByteSlice(buf, x.CapeData),
		WriteString(buf, x.SkinGeometryName),
		WriteByteSlice(buf, x.SkinGeometry),
		WriteString(buf, x.XUID),
		WriteString(buf, x.PlatformChatID),
	)
}

// PlayerAddEntry reads a PlayerListEntry x from Buffer buf in a way that adds a player to the list.
func PlayerAddEntry(buf *bytes.Buffer, x *PlayerListEntry) error {
	return chainErr(
		UUID(buf, &x.UUID),
		Varint64(buf, &x.EntityUniqueID),
		String(buf, &x.Username),
		String(buf, &x.SkinID),
		ByteSlice(buf, &x.SkinData),
		ByteSlice(buf, &x.CapeData),
		String(buf, &x.SkinGeometryName),
		ByteSlice(buf, &x.SkinGeometry),
		String(buf, &x.XUID),
		String(buf, &x.PlatformChatID),
	)
}

// WritePlayerRemoveEntry writes a PlayerListEntry x to Buffer buf in a way that removes a player from the
// list.
func WritePlayerRemoveEntry(buf *bytes.Buffer, x PlayerListEntry) error {
	return WriteUUID(buf, x.UUID)
}

// PlayerRemoveEntry reads a PlayerListEntry x from Buffer buf in a way that removes a player from the list.
func PlayerRemoveEntry(buf *bytes.Buffer, x *PlayerListEntry) error {
	return UUID(buf, &x.UUID)
}
