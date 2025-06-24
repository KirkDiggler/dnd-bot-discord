package dice_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockRoller_Roll(t *testing.T) {
	tests := []struct {
		name        string
		setupRolls  []int
		count       int
		sides       int
		bonus       int
		wantTotal   int
		wantRolls   []int
		wantErr     bool
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
			roller := dice.NewMockRoller()
			roller.SetRolls(tt.setupRolls)

			total, rolls, err := roller.Roll(tt.count, tt.sides, tt.bonus)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.wantRolls, rolls)
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
			roller := dice.NewMockRoller()
			roller.SetRolls(tt.setupRolls)

			total, roll, err := roller.RollWithAdvantage(tt.sides, tt.bonus)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.wantRoll, roll)
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
			roller := dice.NewMockRoller()
			roller.SetRolls(tt.setupRolls)

			total, roll, err := roller.RollWithDisadvantage(tt.sides, tt.bonus)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, total)
			assert.Equal(t, tt.wantRoll, roll)
		})
	}
}

func TestMockRoller_SequentialRolls(t *testing.T) {
	roller := dice.NewMockRoller()
	roller.SetRolls([]int{20, 1, 15, 8})

	// First roll - critical
	total, rolls, err := roller.Roll(1, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, total)
	assert.Equal(t, []int{20}, rolls)

	// Second roll - critical miss
	total, rolls, err = roller.Roll(1, 20, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, []int{1}, rolls)

	// Third roll - normal hit
	total, rolls, err = roller.Roll(1, 20, 5)
	require.NoError(t, err)
	assert.Equal(t, 20, total) // 15+5
	assert.Equal(t, []int{15}, rolls)

	// Fourth roll - damage
	total, rolls, err = roller.Roll(1, 8, 3)
	require.NoError(t, err)
	assert.Equal(t, 11, total) // 8+3
	assert.Equal(t, []int{8}, rolls)

	// Fifth roll should error - no more rolls
	_, _, err = roller.Roll(1, 20, 0)
	assert.Error(t, err)
}

func TestRandomRoller_BasicFunctionality(t *testing.T) {
	// Just verify the random roller doesn't crash
	// We can't test specific values since they're random
	roller := dice.NewRandomRoller()

	// Test basic roll
	total, rolls, err := roller.Roll(2, 6, 3)
	require.NoError(t, err)
	assert.Len(t, rolls, 2)
	assert.GreaterOrEqual(t, total, 5)  // minimum: 1+1+3
	assert.LessOrEqual(t, total, 15)    // maximum: 6+6+3

	// Test advantage
	total, roll, err := roller.RollWithAdvantage(20, 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 3)  // minimum: 1+2
	assert.LessOrEqual(t, total, 22)    // maximum: 20+2
	assert.GreaterOrEqual(t, roll, 1)
	assert.LessOrEqual(t, roll, 20)

	// Test disadvantage
	total, roll, err = roller.RollWithDisadvantage(20, 2)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 3)  // minimum: 1+2
	assert.LessOrEqual(t, total, 22)    // maximum: 20+2
	assert.GreaterOrEqual(t, roll, 1)
	assert.LessOrEqual(t, roll, 20)
}