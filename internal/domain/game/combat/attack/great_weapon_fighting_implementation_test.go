package attack

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollAttackWithFightingStyle_GreatWeaponFighting(t *testing.T) {
	tests := []struct {
		name               string
		rolls              []int
		expectedDamage     int
		expectedRerolls    int
		expectedRerollInfo []DieReroll
		description        string
	}{
		{
			name: "no rerolls needed",
			rolls: []int{
				18,   // Attack
				5, 6, // Damage dice (2d6)
			},
			expectedDamage:  11 + 4, // 5 + 6 + 4 damage bonus
			expectedRerolls: 0,
			description:     "Rolls above 2 are kept",
		},
		{
			name: "reroll single 1",
			rolls: []int{
				15, // Attack
				1,  // First die (reroll)
				4,  // Second die (keep)
				6,  // Reroll of first die
			},
			expectedDamage:  10 + 4, // 6 (rerolled) + 4 + 4 damage bonus
			expectedRerolls: 1,
			expectedRerollInfo: []DieReroll{
				{OriginalRoll: 1, NewRoll: 6, Position: 0},
			},
			description: "Only dice showing 1 or 2 are rerolled",
		},
		{
			name: "reroll both dice",
			rolls: []int{
				12, // Attack
				1,  // First die (reroll)
				2,  // Second die (reroll)
				5,  // Reroll of first
				6,  // Reroll of second
			},
			expectedDamage:  11 + 4, // 5 + 6 + 4 damage bonus
			expectedRerolls: 2,
			expectedRerollInfo: []DieReroll{
				{OriginalRoll: 1, NewRoll: 5, Position: 0},
				{OriginalRoll: 2, NewRoll: 6, Position: 1},
			},
			description: "Both 1s and 2s are rerolled",
		},
		{
			name: "reroll into another low roll",
			rolls: []int{
				10, // Attack
				1,  // First die (reroll)
				5,  // Second die (keep)
				2,  // Reroll gets another 2 (keep it, no second reroll)
			},
			expectedDamage:  7 + 4, // 2 (kept after one reroll) + 5 + 4 damage bonus
			expectedRerolls: 1,
			expectedRerollInfo: []DieReroll{
				{OriginalRoll: 1, NewRoll: 2, Position: 0},
			},
			description: "Each die can only be rerolled once",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoller := mockdice.NewManualMockRoller()
			mockRoller.SetRolls(tt.rolls)

			dmg := &damage.Damage{
				DiceCount:  2,
				DiceSize:   6,
				DamageType: damage.TypeSlashing,
			}

			result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "great_weapon")
			require.NoError(t, err)

			assert.Equal(t, tt.expectedDamage, result.DamageRoll, tt.description)
			assert.Equal(t, tt.expectedRerolls, len(result.RerollInfo), "Number of rerolls should match")

			// Check reroll information
			for i, expectedReroll := range tt.expectedRerollInfo {
				if i < len(result.RerollInfo) {
					actual := result.RerollInfo[i]
					assert.Equal(t, expectedReroll.OriginalRoll, actual.OriginalRoll, "Original roll should match")
					assert.Equal(t, expectedReroll.NewRoll, actual.NewRoll, "New roll should match")
					assert.Equal(t, expectedReroll.Position, actual.Position, "Position should match")
				}
			}
		})
	}
}

func TestRollAttackWithFightingStyle_CriticalHitWithGreatWeapon(t *testing.T) {
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		20, // Natural 20!
		1,  // First regular die (reroll)
		2,  // Second regular die (reroll)
		4,  // Reroll of first
		5,  // Reroll of second
		2,  // First crit die (reroll)
		6,  // Second crit die (keep)
		3,  // Reroll of first crit die
	})

	dmg := &damage.Damage{
		DiceCount:  2,
		DiceSize:   6,
		DamageType: damage.TypeSlashing,
	}

	result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "great_weapon")
	require.NoError(t, err)

	// Should be: 4 + 5 + 3 + 6 + 4 bonus = 22
	assert.Equal(t, 22, result.DamageRoll, "Critical hit damage with GWF")
	assert.Equal(t, 3, len(result.RerollInfo), "Should have 3 rerolls (2 regular + 1 crit)")

	// Check that critical hit flag is set
	assert.True(t, result.AttackResult.IsCrit, "Should be marked as critical hit")
}

func TestRollAttackWithFightingStyle_NoFightingStyle(t *testing.T) {
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		15, // Attack
		1,  // First die (no reroll without GWF)
		2,  // Second die (no reroll without GWF)
	})

	dmg := &damage.Damage{
		DiceCount:  2,
		DiceSize:   6,
		DamageType: damage.TypeSlashing,
	}

	result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "")
	require.NoError(t, err)

	// Should be: 1 + 2 + 4 bonus = 7 (no rerolls)
	assert.Equal(t, 7, result.DamageRoll, "No rerolls without GWF")
	assert.Equal(t, 0, len(result.RerollInfo), "Should have no reroll info")
}

func TestRollAttackWithFightingStyle_DifferentFightingStyle(t *testing.T) {
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		15, // Attack
		1,  // First die (no reroll for defense style)
		2,  // Second die (no reroll for defense style)
	})

	dmg := &damage.Damage{
		DiceCount:  2,
		DiceSize:   6,
		DamageType: damage.TypeSlashing,
	}

	result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "defense")
	require.NoError(t, err)

	// Should be: 1 + 2 + 4 bonus = 7 (no rerolls)
	assert.Equal(t, 7, result.DamageRoll, "No rerolls for other fighting styles")
	assert.Equal(t, 0, len(result.RerollInfo), "Should have no reroll info")
}

func TestRollAttackWithFightingStyle_SingleDieWeapon(t *testing.T) {
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		15, // Attack
		1,  // Die roll (reroll)
		8,  // Reroll
	})

	dmg := &damage.Damage{
		DiceCount:  1,
		DiceSize:   12,
		DamageType: damage.TypeSlashing,
	}

	result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "great_weapon")
	require.NoError(t, err)

	// Should be: 8 + 4 bonus = 12
	assert.Equal(t, 12, result.DamageRoll, "Single die reroll works")
	assert.Equal(t, 1, len(result.RerollInfo), "Should have one reroll")
	assert.Equal(t, 1, result.RerollInfo[0].OriginalRoll, "Original roll should be 1")
	assert.Equal(t, 8, result.RerollInfo[0].NewRoll, "New roll should be 8")
}

func TestGreatWeaponFightingDisplay(t *testing.T) {
	// This test documents how the reroll information should be displayed
	// The UI should format rerolls as: [4, ~~2~~ 5] with strikethrough

	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		15, // Attack
		4,  // First die (keep)
		2,  // Second die (reroll)
		5,  // Reroll of second die
	})

	dmg := &damage.Damage{
		DiceCount:  2,
		DiceSize:   6,
		DamageType: damage.TypeSlashing,
	}

	result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "great_weapon")
	require.NoError(t, err)

	// Verify the data needed for display formatting
	assert.Equal(t, []int{4, 5}, result.AllDamageRolls, "Final rolls should be 4, 5")
	assert.Equal(t, 1, len(result.RerollInfo), "Should have one reroll")

	reroll := result.RerollInfo[0]
	assert.Equal(t, 2, reroll.OriginalRoll, "Original roll was 2")
	assert.Equal(t, 5, reroll.NewRoll, "New roll is 5")
	assert.Equal(t, 1, reroll.Position, "Position is 1 (second die)")

	// The UI layer should use this information to display:
	// "Damage rolls: [4, ~~2~~ 5] = 9 + 4 = 13 slashing damage"
}
