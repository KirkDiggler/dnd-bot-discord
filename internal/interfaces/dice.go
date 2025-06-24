package interfaces

// DiceRoller provides an interface for rolling dice
// This allows us to inject different implementations for testing
type DiceRoller interface {
	// Roll rolls a number of dice with the given sides and adds a bonus
	// Returns the total and individual rolls
	Roll(count, sides, bonus int) (total int, rolls []int, err error)

	// RollWithAdvantage rolls with advantage (roll twice, take higher)
	RollWithAdvantage(sides, bonus int) (total int, roll int, err error)

	// RollWithDisadvantage rolls with disadvantage (roll twice, take lower)
	RollWithDisadvantage(sides, bonus int) (total int, roll int, err error)
}

// RollResult contains detailed information about a dice roll
type RollResult struct {
	Total int   // Sum of all dice plus bonus
	Rolls []int // Individual die results
	Bonus int   // Bonus applied
	Count int   // Number of dice rolled
	Sides int   // Number of sides on each die
}

// DiceRollerV2 provides a more detailed interface for dice rolling
// This will eventually replace DiceRoller
type DiceRollerV2 interface {
	// Roll rolls dice and returns detailed results
	Roll(count, sides, bonus int) (*RollResult, error)

	// RollWithAdvantage rolls with advantage
	RollWithAdvantage(sides, bonus int) (*RollResult, error)

	// RollWithDisadvantage rolls with disadvantage
	RollWithDisadvantage(sides, bonus int) (*RollResult, error)
}
