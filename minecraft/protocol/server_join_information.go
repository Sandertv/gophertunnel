package protocol

import "github.com/google/uuid"

// GatheringJoinInfo contains information about the gathering (experience) the player is joining.
type GatheringJoinInfo struct {
	// ExperienceID is the UUID of the experience.
	ExperienceID uuid.UUID
	// ExperienceName is the name of the experience.
	ExperienceName string
	// ExperienceWorldID is the UUID of the experience world.
	ExperienceWorldID uuid.UUID
	// ExperienceWorldName is the world name of the experience.
	ExperienceWorldName string
	// CreatorID is the ID of the creator.
	CreatorID string
	// UnknownUUID1 is an unknown UUID field.
	UnknownUUID1 uuid.UUID
	// UnknownUUID2 is an unknown UUID field.
	UnknownUUID2 uuid.UUID
	// ServerID is the server identifier.
	ServerID string
}

// Marshal encodes/decodes a GatheringJoinInfo.
func (x *GatheringJoinInfo) Marshal(r IO) {
	r.UUID(&x.ExperienceID)
	r.String(&x.ExperienceName)
	r.UUID(&x.ExperienceWorldID)
	r.String(&x.ExperienceWorldName)
	r.String(&x.CreatorID)
	r.UUID(&x.UnknownUUID1)
	r.UUID(&x.UnknownUUID2)
	r.String(&x.ServerID)
}

// StoreEntryPointInfo contains information about the store entry point.
type StoreEntryPointInfo struct {
	// StoreID is the store identifier.
	StoreID string
	// StoreName is the store name.
	StoreName string
}

// Marshal encodes/decodes a StoreEntryPointInfo.
func (x *StoreEntryPointInfo) Marshal(r IO) {
	r.String(&x.StoreID)
	r.String(&x.StoreName)
}

// PresenceInfo contains presence information about the experience.
type PresenceInfo struct {
	// ExperienceName is the name of the experience.
	ExperienceName string
	// WorldName is the name of the world.
	WorldName string
}

// Marshal encodes/decodes a PresenceInfo.
func (x *PresenceInfo) Marshal(r IO) {
	r.String(&x.ExperienceName)
	r.String(&x.WorldName)
}

// ServerJoinInformation contains optional information about the server the player is joining.
type ServerJoinInformation struct {
	// GatheringJoinInfo is optional information about the gathering being joined.
	GatheringJoinInfo Optional[GatheringJoinInfo]
	// StoreEntryPointInfo is optional information about the store entry point.
	StoreEntryPointInfo Optional[StoreEntryPointInfo]
	// PresenceInfo is optional presence information.
	PresenceInfo Optional[PresenceInfo]
}

// Marshal encodes/decodes a ServerJoinInformation.
func (x *ServerJoinInformation) Marshal(r IO) {
	OptionalMarshaler(r, &x.GatheringJoinInfo)
	OptionalMarshaler(r, &x.StoreEntryPointInfo)
	OptionalMarshaler(r, &x.PresenceInfo)
}
