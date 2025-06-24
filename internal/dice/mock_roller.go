package dice

import (
	"fmt"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/interfaces"
)

// MockRoller implements DiceRoller for testing with predetermined results
type MockRoller struct {
	mu        sync.Mutex
	rolls     []int
	rollIndex int
}

// NewMockRoller creates a new mock dice roller
func NewMockRoller() *MockRoller {
	return &MockRoller{
		rolls: []int{},
	}
}

// SetNextRoll sets the next roll result
func (m *MockRoller) SetNextRoll(roll int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = append(m.rolls, roll)
}

// SetRolls sets multiple roll results
func (m *MockRoller) SetRolls(rolls []int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = rolls
	m.rollIndex = 0
}

// Reset clears all rolls and resets the index
func (m *MockRoller) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rolls = []int{}
	m.rollIndex = 0
}

// getNextRoll returns the next predetermined roll
func (m *MockRoller) getNextRoll() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.rollIndex >= len(m.rolls) {
		return 0, fmt.Errorf("no more predetermined rolls available (used %d of %d)", m.rollIndex, len(m.rolls))
	}

	roll := m.rolls[m.rollIndex]
	m.rollIndex++
	return roll, nil
}

// Roll implements DiceRoller.Roll
func (m *MockRoller) Roll(count, sides, bonus int) (total int, rolls []int, err error) {
	rolls = make([]int, count)
	total = bonus

	for i := 0; i < count; i++ {
		roll, err := m.getNextRoll()
		if err != nil {
			return 0, nil, err
		}
		if roll < 1 || roll > sides {
			return 0, nil, fmt.Errorf("invalid roll %d for d%d", roll, sides)
		}
		rolls[i] = roll
		total += roll
	}

	return total, rolls, nil
}

// RollWithAdvantage implements DiceRoller.RollWithAdvantage
func (m *MockRoller) RollWithAdvantage(sides, bonus int) (total int, roll int, err error) {
	roll1, err := m.getNextRoll()
	if err != nil {
		return 0, 0, err
	}

	roll2, err := m.getNextRoll()
	if err != nil {
		return 0, 0, err
	}

	if roll1 < 1 || roll1 > sides || roll2 < 1 || roll2 > sides {
		return 0, 0, fmt.Errorf("invalid rolls %d,%d for d%d", roll1, roll2, sides)
	}

	// Take the higher roll
	higherRoll := roll1
	if roll2 > higherRoll {
		higherRoll = roll2
	}

	return higherRoll + bonus, higherRoll, nil
}

// RollWithDisadvantage implements DiceRoller.RollWithDisadvantage
func (m *MockRoller) RollWithDisadvantage(sides, bonus int) (total int, roll int, err error) {
	roll1, err := m.getNextRoll()
	if err != nil {
		return 0, 0, err
	}

	roll2, err := m.getNextRoll()
	if err != nil {
		return 0, 0, err
	}

	if roll1 < 1 || roll1 > sides || roll2 < 1 || roll2 > sides {
		return 0, 0, fmt.Errorf("invalid rolls %d,%d for d%d", roll1, roll2, sides)
	}

	// Take the lower roll
	lowerRoll := roll1
	if roll2 < lowerRoll {
		lowerRoll = roll2
	}

	return lowerRoll + bonus, lowerRoll, nil
}

// MockRollerV2 implements DiceRollerV2 for testing
type MockRollerV2 struct {
	*MockRoller
}

// NewMockRollerV2 creates a new mock dice roller with detailed results
func NewMockRollerV2() *MockRollerV2 {
	return &MockRollerV2{
		MockRoller: NewMockRoller(),
	}
}

// Roll implements DiceRollerV2.Roll
func (m *MockRollerV2) Roll(count, sides, bonus int) (*interfaces.RollResult, error) {
	total, rolls, err := m.MockRoller.Roll(count, sides, bonus)
	if err != nil {
		return nil, err
	}

	return &interfaces.RollResult{
		Total: total,
		Rolls: rolls,
		Bonus: bonus,
		Count: count,
		Sides: sides,
	}, nil
}

// RollWithAdvantage implements DiceRollerV2.RollWithAdvantage
func (m *MockRollerV2) RollWithAdvantage(sides, bonus int) (*interfaces.RollResult, error) {
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
	lowerRoll := roll2
	if roll2 > higherRoll {
		higherRoll = roll2
		lowerRoll = roll1
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
func (m *MockRollerV2) RollWithDisadvantage(sides, bonus int) (*interfaces.RollResult, error) {
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
	higherRoll := roll2
	if roll2 < lowerRoll {
		lowerRoll = roll2
		higherRoll = roll1
	}

	return &interfaces.RollResult{
		Total: lowerRoll + bonus,
		Rolls: []int{lowerRoll, higherRoll}, // Show both rolls
		Bonus: bonus,
		Count: 1,
		Sides: sides,
	}, nil
}
