package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	PlayerListActionAdd = iota
	PlayerListActionRemove
)

// PlayerList is sent by the server to update the client-side player list in the in-game menu screen. It shows
// the icon of each player if the correct XUID is written in the packet.
// Sending the PlayerList packet is obligatory when sending an AddPlayer packet. The added player will not
// show up to a client if it has not been added to the player list, because several properties of the player
// are obtained from the player list, such as the skin.
type PlayerList struct {
	// Action is the action to execute upon the player list. The entries that follow specify which entries are
	// added or removed from the player list.
	Action byte
	// Entries is a list of all player list entries that should be added/removed from the player list,
	// depending on the Action set.
	Entries []PlayerListEntry
}

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
	SkinData string
	// CapeData is a byte slice of 64*32*4 bytes. It is a RGBA ordered byte representation of the cape
	// colours, much like the SkinData.
	CapeData string
	// SkinGeometryName is the geometry name of the skin geometry above. This name must be equal to one of the
	// outer names found in the SkinGeometry, so that the client can find the correct geometry data.
	SkinGeometryName string
	// SkinGeometry is a base64 JSON encoded structure of the geometry data of a skin, containing properties
	// such as bones, uv, pivot etc.
	SkinGeometry string
	// XUID is the XBOX Live user ID of the player, which will remain consistent as long as the player is
	// logged in with the XBOX Live account.
	XUID string
	// PlatformChatID is an identifier only set for particular platforms when chatting (presumably only for
	// Nintendo Switch). It is otherwise an empty string, and is used to decide which players are able to
	// chat with each other.
	PlatformChatID string
}

// ID ...
func (*PlayerList) ID() uint32 {
	return IDPlayerList
}

// Marshal ...
func (pk *PlayerList) Marshal(buf *bytes.Buffer) {
	_ = binary.Write(buf, binary.LittleEndian, pk.Action)
	_ = protocol.WriteVaruint32(buf, uint32(len(pk.Entries)))
	for _, entry := range pk.Entries {
		switch pk.Action {
		case PlayerListActionAdd:
			_ = protocol.WriteUUID(buf, entry.UUID)
			_ = protocol.WriteVarint64(buf, entry.EntityUniqueID)
			_ = protocol.WriteString(buf, entry.Username)
			_ = protocol.WriteString(buf, entry.SkinID)
			_ = protocol.WriteString(buf, entry.SkinData)
			_ = protocol.WriteString(buf, entry.CapeData)
			_ = protocol.WriteString(buf, entry.SkinGeometryName)
			_ = protocol.WriteString(buf, entry.SkinGeometry)
			_ = protocol.WriteString(buf, entry.XUID)
			_ = protocol.WriteString(buf, entry.PlatformChatID)
		case PlayerListActionRemove:
			_ = protocol.WriteUUID(buf, entry.UUID)
		default:
			panic(fmt.Sprintf("invalid player list action type %v", pk.Action))
		}
	}
}

// Unmarshal ...
func (pk *PlayerList) Unmarshal(buf *bytes.Buffer) error {
	var count uint32
	if err := chainErr(
		binary.Read(buf, binary.LittleEndian, &pk.Action),
		protocol.Varuint32(buf, &count),
	); err != nil {
		return err
	}
	pk.Entries = make([]PlayerListEntry, count)
	for i := uint32(0); i < count; i++ {
		switch pk.Action {
		case PlayerListActionAdd:
			if err := chainErr(
				protocol.UUID(buf, &pk.Entries[i].UUID),
				protocol.Varint64(buf, &pk.Entries[i].EntityUniqueID),
				protocol.String(buf, &pk.Entries[i].Username),
				protocol.String(buf, &pk.Entries[i].SkinID),
				protocol.String(buf, &pk.Entries[i].SkinData),
				protocol.String(buf, &pk.Entries[i].CapeData),
				protocol.String(buf, &pk.Entries[i].SkinGeometryName),
				protocol.String(buf, &pk.Entries[i].SkinGeometry),
				protocol.String(buf, &pk.Entries[i].XUID),
				protocol.String(buf, &pk.Entries[i].PlatformChatID),
			); err != nil {
				return err
			}
		case PlayerListActionRemove:
			if err := protocol.UUID(buf, &pk.Entries[i].UUID); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown player list action type %v", pk.Action)
		}
	}
	return nil
}
