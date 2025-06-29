package dice

//go:generate mockgen -destination=mock/mock_roller.go -package=mockdice -source=roller.go

// Roller provides an interface for rolling dice
// This allows us to inject different implementations for testing
type Roller interface {
	// Roll rolls a number of dice with the given sides and adds a bonus
	Roll(count, sides, bonus int) (*RollResult, error)

	// RollWithAdvantage rolls with advantage (roll twice, take higher)
	RollWithAdvantage(sides, bonus int) (*RollResult, error)

	// RollWithDisadvantage rolls with disadvantage (roll twice, take lower)
	RollWithDisadvantage(sides, bonus int) (*RollResult, error)
}
