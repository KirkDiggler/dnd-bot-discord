package dice

// randomRoller implements Roller using the existing dice package
type randomRoller struct{}

// NewRandomRoller creates a new random dice roller
func NewRandomRoller() Roller {
	return &randomRoller{}
}

// Roll implements Roller.Roll
func (r *randomRoller) Roll(count, sides, bonus int) (*RollResult, error) {
	result, err := Roll(count, sides, bonus)
	if err != nil {
		return nil, err
	}

	// Convert old RollResult to new format
	rollResult := &RollResult{
		Total:    result.Total,
		Rolls:    result.Rolls,
		Bonus:    bonus,
		Count:    count,
		Sides:    sides,
		RawTotal: result.Total - bonus,
	}

	// Check for crit/fumble on d20
	if count == 1 && sides == 20 && len(result.Rolls) > 0 {
		rollResult.IsCrit = result.Rolls[0] == 20
		rollResult.IsFumble = result.Rolls[0] == 1
	}

	return rollResult, nil
}

// RollWithAdvantage implements Roller.RollWithAdvantage
func (r *randomRoller) RollWithAdvantage(sides, bonus int) (*RollResult, error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}

	result2, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}

	// Take the higher roll
	roll1 := result1.Rolls[0]
	roll2 := result2.Rolls[0]
	higherRoll := roll1
	if roll2 > roll1 {
		higherRoll = roll2
	}

	result := &RollResult{
		Total:    higherRoll + bonus,
		Rolls:    []int{roll1, roll2}, // Show both rolls
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

// RollWithDisadvantage implements Roller.RollWithDisadvantage
func (r *randomRoller) RollWithDisadvantage(sides, bonus int) (*RollResult, error) {
	result1, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}

	result2, err := Roll(1, sides, 0)
	if err != nil {
		return nil, err
	}

	// Take the lower roll
	roll1 := result1.Rolls[0]
	roll2 := result2.Rolls[0]
	lowerRoll := roll1
	if roll2 < roll1 {
		lowerRoll = roll2
	}

	result := &RollResult{
		Total:    lowerRoll + bonus,
		Rolls:    []int{roll1, roll2}, // Show both rolls
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
