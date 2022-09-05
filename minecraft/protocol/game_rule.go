package protocol

// GameRule contains game rule data.
type GameRule struct {
	// Name is the name of the game rule.
	Name string
	// CanBeModifiedByPlayer specifies if the game rule can be modified by the player through the in-game UI.
	CanBeModifiedByPlayer bool
	// Value is the new value of the game rule. This is either a bool, uint32 or float32.
	Value any
}
