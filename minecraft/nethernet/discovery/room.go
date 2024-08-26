package discovery

import (
	"github.com/sandertv/gophertunnel/minecraft/room"
)

func (l *Listener) Announce(status room.Status) error {
	l.ServerData(statusToServerData(status))
	return nil
}

func statusToServerData(status room.Status) *ServerData {
	return &ServerData{
		Version:        0x2,
		ServerName:     status.HostName,
		LevelName:      status.WorldName,
		GameType:       worldTypeToGameType(status.WorldType),
		PlayerCount:    int32(status.MemberCount),
		MaxPlayerCount: int32(status.MaxMemberCount),
		IsEditorWorld:  status.IsEditorWorld,
		TransportLayer: status.TransportLayer,
	}
}

func serverDataToStatus(d *ServerData) room.Status {
	return room.Status{
		HostName:       d.ServerName,
		WorldName:      d.LevelName,
		WorldType:      gameTypeToWorldType(d.GameType),
		MemberCount:    uint32(d.PlayerCount),
		MaxMemberCount: uint32(d.MaxPlayerCount),
		IsEditorWorld:  d.IsEditorWorld,
		TransportLayer: d.TransportLayer,
	}
}

func gameTypeToWorldType(typ int32) string {
	switch typ {
	case 2:
		return room.WorldTypeCreative
	default:
		return room.WorldTypeCreative
	}
}

func worldTypeToGameType(typ string) int32 {
	switch typ {
	case room.WorldTypeCreative:
		return 2
	default:
		return 2
	}
}
