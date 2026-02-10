package protocol

// GatheringJoinInfo contains information about the gathering (experience) the player is joining.
type GatheringJoinInfo struct {
	// ExperienceID is the ID of the experience.
	ExperienceID string
	// ExperienceName is the name of the experience.
	ExperienceName string
	// ExperienceWorldID is the world ID of the experience.
	ExperienceWorldID string
	// ExperienceWorldName is the world name of the experience.
	ExperienceWorldName string
	// CreatorID is the ID of the creator.
	CreatorID string
	// StoreID is the store ID.
	StoreID string
}

// Marshal encodes/decodes a GatheringJoinInfo.
func (x *GatheringJoinInfo) Marshal(r IO) {
	r.String(&x.ExperienceID)
	r.String(&x.ExperienceName)
	r.String(&x.ExperienceWorldID)
	r.String(&x.ExperienceWorldName)
	r.String(&x.CreatorID)
	r.String(&x.StoreID)
}

// ServerJoinInformation contains optional information about the server the player is joining.
type ServerJoinInformation struct {
	// GatheringJoinInfo is optional information about the gathering being joined.
	GatheringJoinInfo Optional[GatheringJoinInfo]
}

// Marshal encodes/decodes a ServerJoinInformation.
func (x *ServerJoinInformation) Marshal(r IO) {
	OptionalMarshaler(r, &x.GatheringJoinInfo)
}
