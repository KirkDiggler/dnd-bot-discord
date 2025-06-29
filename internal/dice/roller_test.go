package dice_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRoller_Roll(t *testing.T) {
	tests := []struct {
		name       string
		setupRolls []int
		count      int
		sides      int
		bonus      int
		wantTotal  int
		wantRolls  []int
		wantErr    bool
	}{
		{
			name:       "single d20 roll",
			setupRolls: []int{15},
			count:      1,
			sides:      20,
			bonus:      0,
			wantTotal:  15,
			wantRolls:  []int{15},
		},
		{
			name:       "2d6+3",
			setupRolls: []int{4, 5},
			count:      2,
			sides:      6,
			bonus:      3,
			wantTotal:  12, // 4+5+3
			wantRolls:  []int{4, 5},
		},
		{
			name:       "critical hit d20",
			setupRolls: []int{20},
			count:      1,
			sides:      20,
			bonus:      5,
			wantTotal:  25,
			wantRolls:  []int{20},
		},
		{
			name:       "not enough rolls",
			setupRolls: []int{10},
			count:      2,
			sides:      6,
			bonus:      0,
			wantErr:    true,
		},
		{
			name:       "invalid roll for die size",
			setupRolls: []int{7},
			count:      1,
			sides:      6,
			bonus:      0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := mockdice.NewManualMockRoller()
			roller.SetRolls(tt.setupRolls)

			result, err := roller.Roll(tt.count, tt.sides, tt.bonus)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, result.Total)
			assert.Equal(t, tt.wantRolls, result.Rolls)
		})
	}
}

func TestMockRoller_RollWithAdvantage(t *testing.T) {
	tests := []struct {
		name       string
		setupRolls []int
		sides      int
		bonus      int
		wantTotal  int
		wantRoll   int
	}{
		{
			name:       "advantage takes higher",
			setupRolls: []int{10, 15},
			sides:      20,
			bonus:      3,
			wantTotal:  18, // 15+3
			wantRoll:   15,
		},
		{
			name:       "advantage with same rolls",
			setupRolls: []int{12, 12},
			sides:      20,
			bonus:      0,
			wantTotal:  12,
			wantRoll:   12,
		},
		{
			name:       "advantage second roll higher",
			setupRolls: []int{8, 17},
			sides:      20,
			bonus:      2,
			wantTotal:  19, // 17+2
			wantRoll:   17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := mockdice.NewManualMockRoller()
			roller.SetRolls(tt.setupRolls)

			result, err := roller.RollWithAdvantage(tt.sides, tt.bonus)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, result.Total)
			assert.Len(t, result.Rolls, 2, "advantage should roll twice")
			// The higher roll should be selected
			higherRoll := result.Rolls[0]
			if result.Rolls[1] > result.Rolls[0] {
				higherRoll = result.Rolls[1]
			}
			assert.Equal(t, tt.wantRoll, higherRoll)
		})
	}
}

func TestMockRoller_RollWithDisadvantage(t *testing.T) {
	tests := []struct {
		name       string
		setupRolls []int
		sides      int
		bonus      int
		wantTotal  int
		wantRoll   int
	}{
		{
			name:       "disadvantage takes lower",
			setupRolls: []int{10, 15},
			sides:      20,
			bonus:      3,
			wantTotal:  13, // 10+3
			wantRoll:   10,
		},
		{
			name:       "disadvantage with same rolls",
			setupRolls: []int{12, 12},
			sides:      20,
			bonus:      0,
			wantTotal:  12,
			wantRoll:   12,
		},
		{
			name:       "disadvantage second roll lower",
			setupRolls: []int{17, 8},
			sides:      20,
			bonus:      2,
			wantTotal:  10, // 8+2
			wantRoll:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roller := mockdice.NewManualMockRoller()
			roller.SetRolls(tt.setupRolls)

			result, err := roller.RollWithDisadvantage(tt.sides, tt.bonus)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, result.Total)
			assert.Len(t, result.Rolls, 2, "disadvantage should roll twice")
			// The lower roll should be selected
			lowerRoll := result.Rolls[0]
			if result.Rolls[1] < result.Rolls[0] {
				lowerRoll = result.Rolls[1]
			}
			assert.Equal(t, tt.wantRoll, lowerRoll)
		})
	}
}

func TestMockRoller_SequentialRolls(t *testing.T) {
	roller := mockdice.NewManualMockRoller()
	roller.SetRolls([]int{20, 1, 15, 8})

	// First roll - critical
	result, err := roller.Roll(1, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, result.Total)
	assert.Equal(t, []int{20}, result.Rolls)

	// Second roll - critical miss
	result, err = roller.Roll(1, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, []int{1}, result.Rolls)

	// Third roll - normal hit
	result, err = roller.Roll(1, 20, 5)
	require.NoError(t, err)
	assert.Equal(t, 20, result.Total) // 15+5
	assert.Equal(t, []int{15}, result.Rolls)

	// Fourth roll - damage
	result, err = roller.Roll(1, 8, 3)
	require.NoError(t, err)
	assert.Equal(t, 11, result.Total) // 8+3
	assert.Equal(t, []int{8}, result.Rolls)

	// Fifth roll should error - no more rolls
	_, err = roller.Roll(1, 20, 0)
	assert.Error(t, err)
}

func TestRandomRoller_BasicFunctionality(t *testing.T) {
	// Just verify the random roller doesn't crash
	// We can't test specific values since they're random
	roller := dice.NewRandomRoller()

	// Test basic roll
	result, err := roller.Roll(2, 6, 3)
	require.NoError(t, err)
	assert.Len(t, result.Rolls, 2)
	assert.GreaterOrEqual(t, result.Total, 5) // minimum: 1+1+3
	assert.LessOrEqual(t, result.Total, 15)   // maximum: 6+6+3

	// Test advantage
	advResult, err := roller.RollWithAdvantage(20, 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, advResult.Total, 3) // minimum: 1+2
	assert.LessOrEqual(t, advResult.Total, 22)   // maximum: 20+2
	assert.Len(t, advResult.Rolls, 2, "advantage should roll twice")
	for _, roll := range advResult.Rolls {
		assert.GreaterOrEqual(t, roll, 1)
		assert.LessOrEqual(t, roll, 20)
	}

	// Test disadvantage
	disResult, err := roller.RollWithDisadvantage(20, 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, disResult.Total, 3) // minimum: 1+2
	assert.LessOrEqual(t, disResult.Total, 22)   // maximum: 20+2
	assert.Len(t, disResult.Rolls, 2, "disadvantage should roll twice")
	for _, roll := range disResult.Rolls {
		assert.GreaterOrEqual(t, roll, 1)
		assert.LessOrEqual(t, roll, 20)
	}
}
