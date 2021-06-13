package constants

type Dimension int32

const (
	Overworld Dimension = iota
	Nether
	End
)

type Difficulty int32

const (
	Peaceful Difficulty = iota
)

type GameMode int32

const (
	Survival GameMode = iota
	Creative
	Adventure
	SurvivalSpectator
	CreativeSpectator
	Default
)

type PlayerPermissionLevel int32

const (
	Visitor PlayerPermissionLevel = iota
	Member
	Operator
	Custom
)
