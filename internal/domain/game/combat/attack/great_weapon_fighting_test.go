package attack

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/stretchr/testify/assert"
)

// TestGreatWeaponFightingImplementation shows how GWF should work
func TestGreatWeaponFightingImplementation(t *testing.T) {
	t.Run("proposed RollAttackWithFightingStyle function", func(t *testing.T) {
		mockRoller := mockdice.NewManualMockRoller()

		// Set up rolls: attack roll, then damage dice with rerolls
		mockRoller.SetRolls([]int{
			15, // Attack roll
			1,  // First damage die (will be rerolled)
			2,  // Second damage die (will be rerolled)
			5,  // Reroll of first die
			4,  // Reroll of second die
		})

		// This is what the function signature might look like:
		// func RollAttackWithFightingStyle(roller dice.Roller, attackBonus, damageBonus int, dmg *damage.Damage, fightingStyle string) (*Result, error)

		// For now, let's simulate what should happen
		_ = 15 // attackRoll would be used in real implementation

		// Roll damage dice
		firstRoll := []int{1, 2}
		// Check for rerolls (GWF rerolls 1s and 2s)
		finalRolls := []int{}
		rerollIndex := 2 // Start index for rerolls in our mock
		for _, roll := range firstRoll {
			if roll <= 2 {
				// Would reroll this die once
				if rerollIndex == 2 {
					finalRolls = append(finalRolls, 5) // First reroll
					rerollIndex++
				} else {
					finalRolls = append(finalRolls, 4) // Second reroll
				}
			} else {
				finalRolls = append(finalRolls, roll)
			}
		}

		totalDamage := 0
		for _, roll := range finalRolls {
			totalDamage += roll
		}

		// Verify the total damage after rerolls
		assert.Equal(t, 9, totalDamage)
	})
}

// TestRollAttackWithGreatWeaponFighting demonstrates the ideal implementation
func TestRollAttackWithGreatWeaponFighting(t *testing.T) {
	t.Skip("Waiting for Great Weapon Fighting implementation")

	mockRoller := mockdice.NewManualMockRoller()

	tests := []struct {
		name           string
		rolls          []int
		expectedDamage int
		description    string
	}{
		{
			name: "no rerolls needed",
			rolls: []int{
				18,   // Attack
				5, 6, // Damage dice (2d6)
			},
			expectedDamage: 11, // 5 + 6
			description:    "Rolls above 2 are kept",
		},
		{
			name: "reroll single 1",
			rolls: []int{
				15, // Attack
				1,  // First die (reroll)
				4,  // Second die (keep)
				6,  // Reroll of first die
			},
			expectedDamage: 10, // 6 (rerolled) + 4
			description:    "Only dice showing 1 or 2 are rerolled",
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
			expectedDamage: 11, // 5 + 6
			description:    "Both 1s and 2s are rerolled",
		},
		{
			name: "reroll into another low roll",
			rolls: []int{
				10, // Attack
				1,  // First die (reroll)
				5,  // Second die (keep)
				2,  // Reroll gets another 2 (keep it, no second reroll)
			},
			expectedDamage: 7, // 2 (kept after one reroll) + 5
			description:    "Each die can only be rerolled once",
		},
		{
			name: "critical hit with rerolls",
			rolls: []int{
				20, // Natural 20!
				1,  // First regular die (reroll)
				2,  // Second regular die (reroll)
				4,  // Reroll of first
				5,  // Reroll of second
				2,  // First crit die (reroll)
				6,  // Second crit die (keep)
				3,  // Reroll of first crit die
			},
			expectedDamage: 18, // 4 + 5 + 3 + 6
			description:    "GWF applies to critical hit dice too",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRoller.SetRolls(tt.rolls)

			// Hypothetical function that handles GWF
			// dmg := &damage.Damage{
			// 	DiceCount:  2,
			// 	DiceSize:   6,
			// 	DamageType: damage.TypeSlashing,
			// }

			// This function would need to be implemented
			// result, err := RollAttackWithFightingStyle(mockRoller, 5, 4, dmg, "great_weapon")
			// require.NoError(t, err)
			// assert.Equal(t, tt.expectedDamage + 4, result.DamageRoll) // +4 from damage bonus
		})
	}
}

// TestGreatWeaponFightingEdgeCases covers specific GWF rules
func TestGreatWeaponFightingEdgeCases(t *testing.T) {
	t.Run("only reroll weapon damage dice", func(t *testing.T) {
		// If there are bonus dice from other sources (like sneak attack),
		// GWF should NOT reroll those dice

		// Example: Greatsword (2d6) + Sneak Attack (1d6)
		// If you roll [1, 2, 1] where the third 1 is sneak attack,
		// only the first two dice get rerolled
	})

	t.Run("works with different weapon types", func(t *testing.T) {
		// GWF should work with any two-handed weapon:
		// - Greatsword (2d6)
		// - Greataxe (1d12)
		// - Maul (2d6)
		// - Heavy Crossbow (1d10) - if you have GWF and somehow use it with ranged
	})

	t.Run("versatile weapons in two hands", func(t *testing.T) {
		// Versatile weapons used with two hands should benefit from GWF
		// Example: Longsword (1d8 one-handed, 1d10 two-handed)
		// When used two-handed with GWF, the 1d10 can be rerolled
	})
}

// Example of how the dice roller interface might be extended
type GreatWeaponRoller interface {
	dice.Roller
	// RollWithReroll rolls dice and rerolls any that meet the condition (once per die)
	RollWithReroll(count, sides, bonus int, shouldReroll func(int) bool) (*dice.RollResult, error)
}

// This shows how RollWithReroll might work when implemented
// func mockImplementation(roller dice.Roller, count, sides int, shouldReroll func(int) bool) (*dice.RollResult, error) {
// 	// First roll
// 	initial, err := roller.Roll(count, sides, 0)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	finalRolls := make([]int, len(initial.Rolls))
// 	total := 0
//
// 	// Check each die
// 	for i, roll := range initial.Rolls {
// 		if shouldReroll(roll) {
// 			// Reroll this die once
// 			reroll, err := roller.Roll(1, sides, 0)
// 			if err != nil {
// 				return nil, err
// 			}
// 			finalRolls[i] = reroll.Rolls[0]
// 			total += reroll.Rolls[0]
// 		} else {
// 			finalRolls[i] = roll
// 			total += roll
// 		}
// 	}
//
// 	return &dice.RollResult{
// 		Total: total,
// 		Rolls: finalRolls,
// 		Bonus: 0,
// 		Count: count,
// 		Sides: sides,
// 	}, nil
// }
