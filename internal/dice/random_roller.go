package dice

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/interfaces"
)

// RandomRoller implements DiceRoller using the existing dice package
type RandomRoller struct{}

// NewRandomRoller creates a new random dice roller
func NewRandomRoller() interfaces.DiceRoller {
	return &RandomRoller{}
}

// Roll implements DiceRoller.Roll
func (r *RandomRoller) Roll(count, sides, bonus int) (total int, rolls []int, err error) {
	result, err := Roll(count, sides, bonus)
	if err != nil {
		return 0, nil, err
	}
	return result.Total, result.Rolls, nil
}

// RollWithAdvantage implements DiceRoller.RollWithAdvantage
func (r *RandomRoller) RollWithAdvantage(sides, bonus int) (total int, roll int, err error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return 0, 0, err
	}
	
	result2, err := Roll(1, sides, 0)
	if err != nil {
		return 0, 0, err
	}
	
	// Take the higher roll
	higherRoll := result1.Rolls[0]
	if result2.Rolls[0] > higherRoll {
		higherRoll = result2.Rolls[0]
	}
	
	return higherRoll + bonus, higherRoll, nil
}

// RollWithDisadvantage implements DiceRoller.RollWithDisadvantage
func (r *RandomRoller) RollWithDisadvantage(sides, bonus int) (total int, roll int, err error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return 0, 0, err
	}
	
	result2, err := Roll(1, sides, 0)
	if err != nil {
		return 0, 0, err
	}
	
	// Take the lower roll
	lowerRoll := result1.Rolls[0]
	if result2.Rolls[0] < lowerRoll {
		lowerRoll = result2.Rolls[0]
	}
	
	return lowerRoll + bonus, lowerRoll, nil
}

// RandomRollerV2 implements DiceRollerV2 using the existing dice package
type RandomRollerV2 struct{}

// NewRandomRollerV2 creates a new random dice roller with detailed results
func NewRandomRollerV2() interfaces.DiceRollerV2 {
	return &RandomRollerV2{}
}

// Roll implements DiceRollerV2.Roll
func (r *RandomRollerV2) Roll(count, sides, bonus int) (*interfaces.RollResult, error) {
	result, err := Roll(count, sides, bonus)
	if err != nil {
		return nil, err
	}
	
	return &interfaces.RollResult{
		Total: result.Total,
		Rolls: result.Rolls,
		Bonus: bonus,
		Count: count,
		Sides: sides,
	}, nil
}

// RollWithAdvantage implements DiceRollerV2.RollWithAdvantage
func (r *RandomRollerV2) RollWithAdvantage(sides, bonus int) (*interfaces.RollResult, error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}
	
	result2, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}
	
	// Take the higher roll
	higherRoll := result1.Rolls[0]
	lowerRoll := result2.Rolls[0]
	if result2.Rolls[0] > higherRoll {
		higherRoll = result2.Rolls[0]
		lowerRoll = result1.Rolls[0]
	}
	
	return &interfaces.RollResult{
		Total: higherRoll + bonus,
		Rolls: []int{higherRoll, lowerRoll}, // Show both rolls
		Bonus: bonus,
		Count: 1,
		Sides: sides,
	}, nil
}

// RollWithDisadvantage implements DiceRollerV2.RollWithDisadvantage
func (r *RandomRollerV2) RollWithDisadvantage(sides, bonus int) (*interfaces.RollResult, error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}
	
	result2, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}
	
	// Take the lower roll
	lowerRoll := result1.Rolls[0]
	higherRoll := result2.Rolls[0]
	if result2.Rolls[0] < lowerRoll {
		lowerRoll = result2.Rolls[0]
		higherRoll = result1.Rolls[0]
	}
	
	return &interfaces.RollResult{
		Total: lowerRoll + bonus,
		Rolls: []int{lowerRoll, higherRoll}, // Show both rolls
		Bonus: bonus,
		Count: 1,
		Sides: sides,
	}, nil
}