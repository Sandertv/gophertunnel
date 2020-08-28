package protocol

import (
	"bytes"
	"encoding/binary"
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
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
	// BuildPlatform is the platform of the player as sent by that player in the Login packet.
	BuildPlatform int32
	// Skin is the skin of the player that should be added to the player list. Once sent here, it will not
	// have to be sent again.
	Skin Skin
	// Teacher is a Minecraft: Education Edition field. It specifies if the player to be added to the player
	// list is a teacher.
	Teacher bool
	// Host specifies if the player that is added to the player list is the host of the game.
	Host bool
}

// WritePlayerAddEntry writes a PlayerListEntry x to Buffer buf in a way that adds the player to the list.
func WritePlayerAddEntry(buf *bytes.Buffer, x PlayerListEntry) error {
	return chainErr(
		WriteUUID(buf, x.UUID),
		WriteVarint64(buf, x.EntityUniqueID),
		WriteString(buf, x.Username),
		WriteString(buf, x.XUID),
		WriteString(buf, x.PlatformChatID),
		binary.Write(buf, binary.LittleEndian, x.BuildPlatform),
		WriteSerialisedSkin(buf, x.Skin),
		binary.Write(buf, binary.LittleEndian, x.Teacher),
		binary.Write(buf, binary.LittleEndian, x.Host),
	)
}

// PlayerAddEntry reads a PlayerListEntry x from Reader r in a way that adds a player to the list.
func PlayerAddEntry(r *Reader, x *PlayerListEntry) {
	r.UUID(&x.UUID)
	r.Varint64(&x.EntityUniqueID)
	r.String(&x.Username)
	r.String(&x.XUID)
	r.String(&x.PlatformChatID)
	r.Int32(&x.BuildPlatform)
	SerialisedSkin(r, &x.Skin)
	r.Bool(&x.Teacher)
	r.Bool(&x.Host)
}

// WritePlayerRemoveEntry writes a PlayerListEntry x to Buffer buf in a way that removes a player from the
// list.
func WritePlayerRemoveEntry(buf *bytes.Buffer, x PlayerListEntry) error {
	return WriteUUID(buf, x.UUID)
}

// PlayerRemoveEntry reads a PlayerListEntry x from Reader r in a way that removes a player from the list.
func PlayerRemoveEntry(r *Reader, x *PlayerListEntry) {
	r.UUID(&x.UUID)
}
