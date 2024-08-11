package mpsd

import (
	"encoding/json"
	"github.com/google/uuid"
)

type MemberDescription struct {
	Constants  *MemberConstants  `json:"constants,omitempty"`
	Properties *MemberProperties `json:"properties,omitempty"`
	Roles      json.RawMessage   `json:"roles,omitempty"`
}

type MemberProperties struct {
	System *MemberPropertiesSystem `json:"system,omitempty"`
	Custom json.RawMessage         `json:"custom,omitempty"`
}

type MemberPropertiesSystem struct {
	Active              bool                                `json:"active,omitempty"`
	Ready               bool                                `json:"ready,omitempty"`
	Connection          uuid.UUID                           `json:"connection,omitempty"`
	Subscription        *MemberPropertiesSystemSubscription `json:"subscription,omitempty"`
	SecureDeviceAddress []byte                              `json:"secureDeviceAddress,omitempty"`
	InitializationGroup []uint32                            `json:"initializationGroup,omitempty"`
	Groups              []string                            `json:"groups,omitempty"`
	Encounters          []string                            `json:"encounters,omitempty"`
	Measurements        json.RawMessage                     `json:"measurements,omitempty"`
	ServerMeasurements  json.RawMessage                     `json:"serverMeasurements,omitempty"`
}

type MemberPropertiesSystemSubscription struct {
	ID          string   `json:"id,omitempty"`
	ChangeTypes []string `json:"changeTypes,omitempty"`
}

const (
	ChangeTypeEverything            = "everything"
	ChangeTypeHost                  = "host"
	ChangeTypeInitialization        = "initialization"
	ChangeTypeMatchmakingStatus     = "matchmakingStatus"
	ChangeTypeMembersList           = "membersList"
	ChangeTypeMembersStatus         = "membersStatus"
	ChangeTypeJoinability           = "joinability"
	ChangeTypeCustomProperty        = "customProperty"
	ChangeTypeMembersCustomProperty = "membersCustomProperty"
)

type MemberConstants struct {
	System *MemberConstantsSystem `json:"system,omitempty"`
	Custom json.RawMessage        `json:"custom,omitempty"`
}

type MemberConstantsSystem struct {
	XUID       string `json:"xuid,omitempty"`
	Initialize bool   `json:"initialize,omitempty"`
}
