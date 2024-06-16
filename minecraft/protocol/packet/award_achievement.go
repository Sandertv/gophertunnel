package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// AwardAchievement is sent by the server to award an achievement to a player.
type AwardAchievement struct {
	// AchievementID is the ID of the achievement that should be awarded to the player. The values for these
	// IDs are currently unknown.
	AchievementID int32
}

// ID ...
func (*AwardAchievement) ID() uint32 {
	return IDAwardAchievement
}

func (pk *AwardAchievement) Marshal(io protocol.IO) {
	io.Int32(&pk.AchievementID)
}
