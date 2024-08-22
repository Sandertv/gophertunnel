package room

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
	MemberCount             uint32       `json:"MemberCount"`
	MaxMemberCount          uint32       `json:"MaxMemberCount"`
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
