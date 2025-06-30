package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonkMartialArts_DamageScaling(t *testing.T) {
	tests := []struct {
		name             string
		level            int
		hasMartialArts   bool
		expectedDiceSize int
		description      string
	}{
		{
			name:             "Level 1 monk uses 1d4",
			level:            1,
			hasMartialArts:   true,
			expectedDiceSize: 4,
			description:      "Monks at levels 1-4 use 1d4 for unarmed strikes",
		},
		{
			name:             "Level 4 monk uses 1d4",
			level:            4,
			hasMartialArts:   true,
			expectedDiceSize: 4,
			description:      "Monks at levels 1-4 use 1d4 for unarmed strikes",
		},
		{
			name:             "Level 5 monk uses 1d6",
			level:            5,
			hasMartialArts:   true,
			expectedDiceSize: 6,
			description:      "Monks at levels 5-10 use 1d6 for unarmed strikes",
		},
		{
			name:             "Level 10 monk uses 1d6",
			level:            10,
			hasMartialArts:   true,
			expectedDiceSize: 6,
			description:      "Monks at levels 5-10 use 1d6 for unarmed strikes",
		},
		{
			name:             "Level 11 monk uses 1d8",
			level:            11,
			hasMartialArts:   true,
			expectedDiceSize: 8,
			description:      "Monks at levels 11-16 use 1d8 for unarmed strikes",
		},
		{
			name:             "Level 16 monk uses 1d8",
			level:            16,
			hasMartialArts:   true,
			expectedDiceSize: 8,
			description:      "Monks at levels 11-16 use 1d8 for unarmed strikes",
		},
		{
			name:             "Level 17 monk uses 1d10",
			level:            17,
			hasMartialArts:   true,
			expectedDiceSize: 10,
			description:      "Monks at levels 17-20 use 1d10 for unarmed strikes",
		},
		{
			name:             "Level 20 monk uses 1d10",
			level:            20,
			hasMartialArts:   true,
			expectedDiceSize: 10,
			description:      "Monks at levels 17-20 use 1d10 for unarmed strikes",
		},
		{
			name:             "Level 20 non-monk uses 1d4",
			level:            20,
			hasMartialArts:   false,
			expectedDiceSize: 4,
			description:      "Non-monks always use 1d4 regardless of level",
		},
		{
			name:             "Level 5 fighter uses 1d4",
			level:            5,
			hasMartialArts:   false,
			expectedDiceSize: 4,
			description:      "Only monks get scaling damage dice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character with specified level and features
			features := []*rulebook.CharacterFeature{}
			if tt.hasMartialArts {
				features = append(features, &rulebook.CharacterFeature{Key: "martial-arts", Name: "Martial Arts"})
			}

			char := &character.Character{
				Level:    tt.level,
				Features: features,
				Attributes: map[character.Attribute]*character.AbilityScore{
					character.AttributeStrength: {
						Score: 14, // +2 bonus
						Bonus: 2,
					},
					character.AttributeDexterity: {
						Score: 16, // +3 bonus
						Bonus: 3,
					},
				},
			}

			// Mock dice roller - we'll check what dice size was requested
			mockRoller := mockdice.NewManualMockRoller()
			// Set rolls: attack d20, damage die (size varies)
			mockRoller.SetRolls([]int{
				15,                      // Attack roll (d20)
				tt.expectedDiceSize / 2, // Damage roll (varies by level)
			})
			char = char.WithDiceRoller(mockRoller)

			// Perform unarmed strike
			result, err := char.improvisedMelee()
			require.NoError(t, err)

			// Verify the damage dice size in the result
			assert.Equal(t, 1, result.WeaponDamage.DiceCount, "Should always roll 1 die")
			assert.Equal(t, tt.expectedDiceSize, result.WeaponDamage.DiceSize, tt.description)

			// Verify damage calculation
			expectedBonus := 3 // DEX bonus for monks
			if !tt.hasMartialArts {
				expectedBonus = 2 // STR bonus for non-monks
			}
			expectedDamage := tt.expectedDiceSize/2 + expectedBonus
			assert.Equal(t, expectedDamage, result.DamageRoll, "Damage should be die roll + ability bonus")
		})
	}
}

func TestMonkMartialArts_DamageScalingWithCombat(t *testing.T) {
	// Test that damage scaling works correctly in different combat scenarios
	tests := []struct {
		name           string
		level          int
		strScore       int
		dexScore       int
		attackRoll     int
		damageRoll     int
		expectedAttack int
		expectedDamage int
		expectedDice   int
	}{
		{
			name:           "Level 5 monk with high DEX",
			level:          5,
			strScore:       10,
			dexScore:       18, // +4 bonus
			attackRoll:     12,
			damageRoll:     5,  // rolling 5 on a d6
			expectedAttack: 16, // 12 + 4
			expectedDamage: 9,  // 5 + 4
			expectedDice:   6,
		},
		{
			name:           "Level 11 monk with equal STR/DEX",
			level:          11,
			strScore:       16, // +3 bonus
			dexScore:       16, // +3 bonus
			attackRoll:     20, // Critical hit!
			damageRoll:     7,  // rolling 7 on a d8
			expectedAttack: 23, // 20 + 3
			expectedDamage: 10, // 7 + 3
			expectedDice:   8,
		},
		{
			name:           "Level 17 monk with higher STR",
			level:          17,
			strScore:       20, // +5 bonus
			dexScore:       14, // +2 bonus
			attackRoll:     1,  // Critical miss!
			damageRoll:     8,  // rolling 8 on a d10
			expectedAttack: 6,  // 1 + 5
			expectedDamage: 13, // 8 + 5
			expectedDice:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monk := &character.Character{
				Level: tt.level,
				Features: []*rulebook.CharacterFeature{
					{Key: "martial-arts", Name: "Martial Arts"},
				},
				Attributes: map[character.Attribute]*character.AbilityScore{
					character.AttributeStrength: {
						Score: tt.strScore,
						Bonus: (tt.strScore - 10) / 2,
					},
					character.AttributeDexterity: {
						Score: tt.dexScore,
						Bonus: (tt.dexScore - 10) / 2,
					},
				},
			}

			mockRoller := mockdice.NewManualMockRoller()
			mockRoller.SetRolls([]int{tt.attackRoll, tt.damageRoll})
			monk = monk.WithDiceRoller(mockRoller)

			result, err := monk.improvisedMelee()
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAttack, result.AttackRoll)
			assert.Equal(t, tt.expectedDamage, result.DamageRoll)
			assert.Equal(t, tt.expectedDice, result.WeaponDamage.DiceSize)
		})
	}
}
