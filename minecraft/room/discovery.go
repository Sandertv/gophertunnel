package room

import (
	"context"
	"github.com/df-mc/go-nethernet/discovery"
)

type DiscoveryAnnouncer struct {
	Listener *discovery.Listener
}

func (a DiscoveryAnnouncer) Announce(_ context.Context, status Status) {
	a.Listener.ServerData(statusToServerData(status))
}

func (a DiscoveryAnnouncer) Close() error {
	return a.Listener.Close()
}

func statusToServerData(status Status) *discovery.ServerData {
	return &discovery.ServerData{
		Version:        0x2,
		ServerName:     status.HostName,
		LevelName:      status.WorldName,
		GameType:       worldTypeToGameType(status.WorldType),
		PlayerCount:    int32(status.MemberCount),
		MaxPlayerCount: int32(status.MaxMemberCount),
		TransportLayer: status.TransportLayer,
	}
}

func worldTypeToGameType(typ string) int32 {
	switch typ {
	case WorldTypeCreative:
		return 2
	default:
		return 2
	}
}
