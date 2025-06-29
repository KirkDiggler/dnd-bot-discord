package mockdice

import (
	"fmt"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
)

// ManualMockRoller implements dice.Roller for testing with predetermined results
type ManualMockRoller struct {
	mu        sync.Mutex
	rolls     []int
	rollIndex int
}

// NewManualMockRoller creates a new mock dice roller
func NewManualMockRoller() *ManualMockRoller {
	return &ManualMockRoller{
		rolls: []int{},
	}
}

// SetNextRoll sets the next roll result
func (m *ManualMockRoller) SetNextRoll(roll int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = append(m.rolls, roll)
}

// SetRolls sets multiple roll results
func (m *ManualMockRoller) SetRolls(rolls []int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = rolls
	m.rollIndex = 0
}

// Reset clears all rolls and resets the index
func (m *ManualMockRoller) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = []int{}
	m.rollIndex = 0
}

// getNextRoll returns the next predetermined roll
func (m *ManualMockRoller) getNextRoll() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rollIndex >= len(m.rolls) {
		return 0, fmt.Errorf("no more predetermined rolls available (used %d of %d)", m.rollIndex, len(m.rolls))
	}

	roll := m.rolls[m.rollIndex]
	m.rollIndex++
	return roll, nil
}

// Roll implements dice.Roller.Roll
func (m *ManualMockRoller) Roll(count, sides, bonus int) (*dice.RollResult, error) {
	rolls := make([]int, count)
	rawTotal := 0

	for i := 0; i < count; i++ {
		roll, err := m.getNextRoll()
		if err != nil {
			return nil, err
		}
		if roll < 1 || roll > sides {
			return nil, fmt.Errorf("invalid roll %d for d%d", roll, sides)
		}
		rolls[i] = roll
		rawTotal += roll
	}

	result := &dice.RollResult{
		Total:    rawTotal + bonus,
		Rolls:    rolls,
		Bonus:    bonus,
		Count:    count,
		Sides:    sides,
		RawTotal: rawTotal,
	}

	// Check for crit/fumble on d20
	if count == 1 && sides == 20 && len(rolls) > 0 {
		result.IsCrit = rolls[0] == 20
		result.IsFumble = rolls[0] == 1
	}

	return result, nil
}

// RollWithAdvantage implements dice.Roller.RollWithAdvantage
func (m *ManualMockRoller) RollWithAdvantage(sides, bonus int) (*dice.RollResult, error) {
	roll1, err := m.getNextRoll()
	if err != nil {
		return nil, err
	}

	roll2, err := m.getNextRoll()
	if err != nil {
		return nil, err
	}

	if roll1 < 1 || roll1 > sides || roll2 < 1 || roll2 > sides {
		return nil, fmt.Errorf("invalid rolls %d,%d for d%d", roll1, roll2, sides)
	}

	// Take the higher roll
	higherRoll := roll1
	if roll2 > higherRoll {
		higherRoll = roll2
	}

	result := &dice.RollResult{
		Total:    higherRoll + bonus,
		Rolls:    []int{roll1, roll2},
		Bonus:    bonus,
		Count:    1,
		Sides:    sides,
		RawTotal: higherRoll,
	}

	// Check for crit/fumble on d20
	if sides == 20 {
		result.IsCrit = higherRoll == 20
		result.IsFumble = higherRoll == 1
	}

	return result, nil
}

// RollWithDisadvantage implements dice.Roller.RollWithDisadvantage
func (m *ManualMockRoller) RollWithDisadvantage(sides, bonus int) (*dice.RollResult, error) {
	roll1, err := m.getNextRoll()
	if err != nil {
		return nil, err
	}

	roll2, err := m.getNextRoll()
	if err != nil {
		return nil, err
	}

	if roll1 < 1 || roll1 > sides || roll2 < 1 || roll2 > sides {
		return nil, fmt.Errorf("invalid rolls %d,%d for d%d", roll1, roll2, sides)
	}

	// Take the lower roll
	lowerRoll := roll1
	if roll2 < lowerRoll {
		lowerRoll = roll2
	}

	result := &dice.RollResult{
		Total:    lowerRoll + bonus,
		Rolls:    []int{roll1, roll2},
		Bonus:    bonus,
		Count:    1,
		Sides:    sides,
		RawTotal: lowerRoll,
	}

	// Check for crit/fumble on d20
	if sides == 20 {
		result.IsCrit = lowerRoll == 20
		result.IsFumble = lowerRoll == 1
	}

	return result, nil
}
