package room

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type Status struct {
	Joinability             string       `json:"Joinability,omitempty"`
	HostName                string       `json:"hostName,omitempty"`
	OwnerID                 string       `json:"ownerId,omitempty"`
	RakNetGUID              string       `json:"rakNetGUID"`
	Version                 string       `json:"version"`
	LevelID                 string       `json:"levelId"`
	WorldName               string       `json:"worldName"`
	WorldType               string       `json:"worldType"`
	Protocol                int32        `json:"protocol"`
	MemberCount             int          `json:"MemberCount"`
	MaxMemberCount          int          `json:"MaxMemberCount"`
	BroadcastSetting        uint32       `json:"BroadcastSetting"`
	LanGame                 bool         `json:"LanGame"`
	IsEditorWorld           bool         `json:"isEditorWorld"`
	TransportLayer          int32        `json:"TransportLayer"`
	WebRTCNetworkID         uint64       `json:"WebRTCNetworkId"`
	OnlineCrossPlatformGame bool         `json:"OnlineCrossPlatformGame"`
	CrossPlayDisabled       bool         `json:"CrossPlayDisabled"`
	TitleID                 int64        `json:"TitleId"`
	SupportedConnections    []Connection `json:"SupportedConnections"`
}

type Connection struct {
	ConnectionType  uint32 `json:"ConnectionType"`
	HostIPAddress   string `json:"HostIpAddress"`
	HostPort        uint16 `json:"HostPort"`
	NetherNetID     uint64 `json:"NetherNetId"`
	WebRTCNetworkID uint64 `json:"WebRTCNetworkId"`
	RakNetGUID      string `json:"RakNetGUID"`
}

const (
	JoinabilityInviteOnly        = "invite_only"
	JoinabilityJoinableByFriends = "joinable_by_friends"
)

const (
	WorldTypeCreative = "Creative"
)

const (
	BroadcastSettingInviteOnly uint32 = iota + 1
	BroadcastSettingFriendsOnly
	BroadcastSettingFriendsOfFriends
)

const (
	TransportLayerRakNet int32 = iota
	_
	TransportLayerNetherNet
)

const (
	ConnectionTypeWebSocketsWebRTCSignaling uint32 = 3
)

type StatusProvider interface {
	RoomStatus() Status
}

func NewStatusProvider(status Status) StatusProvider {
	return statusProvider{status: status}
}

type statusProvider struct{ status Status }

func (p statusProvider) RoomStatus() Status {
	return p.status
}

func DefaultStatus() Status {
	levelID := make([]byte, 8)
	_, _ = rand.Read(levelID)

	return Status{
		Joinability:             JoinabilityJoinableByFriends,
		HostName:                "Gophertunnel",
		Version:                 protocol.CurrentVersion,
		LevelID:                 base64.StdEncoding.EncodeToString(levelID),
		WorldName:               "Room Listener",
		WorldType:               WorldTypeCreative,
		Protocol:                protocol.CurrentProtocol,
		BroadcastSetting:        BroadcastSettingFriendsOfFriends,
		LanGame:                 true,
		TransportLayer:          TransportLayerNetherNet,
		OnlineCrossPlatformGame: true,
		CrossPlayDisabled:       false,
		TitleID:                 0,
	}
}

func NetherNetID(status Status) (uint64, bool) {
	for _, c := range status.SupportedConnections {
		if c.ConnectionType == ConnectionTypeWebSocketsWebRTCSignaling {
			if c.WebRTCNetworkID != 0 {
				return c.WebRTCNetworkID, true
			}
			if c.NetherNetID != 0 {
				return c.NetherNetID, true
			}
		}
	}
	return 0, false
}
