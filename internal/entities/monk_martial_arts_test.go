package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"

	mockdice "github.com/KirkDiggler/dnd-bot-discord/internal/dice/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonkMartialArts_UnarmedStrike(t *testing.T) {
	tests := []struct {
		name           string
		features       []*rulebook.CharacterFeature
		strScore       int
		dexScore       int
		expectedAttack int // Expected attack bonus (not including d20)
		expectedDamage int // Expected damage bonus (not including d4)
		description    string
	}{
		{
			name:           "Monk with higher DEX uses DEX for attack and damage",
			features:       []*rulebook.CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       10, // +0 bonus
			dexScore:       16, // +3 bonus
			expectedAttack: 3,
			expectedDamage: 3,
			description:    "Monks should use DEX when it's higher than STR",
		},
		{
			name:           "Monk with higher STR uses STR for attack and damage",
			features:       []*rulebook.CharacterFeature{{Key: "martial-arts", Name: "Martial Arts"}},
			strScore:       18, // +4 bonus
			dexScore:       14, // +2 bonus
			expectedAttack: 4,
			expectedDamage: 4,
			description:    "Monks can still use STR if it's higher",
		},
		{
			name:           "Non-monk always uses STR",
			features:       []*rulebook.CharacterFeature{}, // No martial arts
			strScore:       10,                             // +0 bonus
			dexScore:       18,                             // +4 bonus
			expectedAttack: 0,
			expectedDamage: 0,
			description:    "Characters without Martial Arts must use STR",
		},
		{
			name:           "Non-monk fighter uses STR even with high DEX",
			features:       []*rulebook.CharacterFeature{{Key: "dueling", Name: "Dueling"}},
			strScore:       12, // +1 bonus
			dexScore:       16, // +3 bonus
			expectedAttack: 1,
			expectedDamage: 1,
			description:    "Other classes can't use DEX for unarmed strikes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character with specified attributes
			char := &character.Character{
				Features: tt.features,
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

			// Mock dice roller
			mockRoller := mockdice.NewManualMockRoller()
			mockRoller.SetRolls([]int{
				15, // Attack roll
				3,  // Damage roll (1d4)
			})
			char = char.WithDiceRoller(mockRoller)

			// Perform unarmed strike
			result, err := char.improvisedMelee()
			require.NoError(t, err)

			// Verify attack and damage calculations
			assert.Equal(t, 15+tt.expectedAttack, result.AttackRoll, "Attack roll should be d20 + ability bonus")
			assert.Equal(t, 3+tt.expectedDamage, result.DamageRoll, "Damage roll should be 1d4 + ability bonus")
			assert.Equal(t, damage.TypeBludgeoning, result.AttackType, "Unarmed strikes deal bludgeoning damage")

			// Verify the damage dice are still 1d4
			assert.Equal(t, 1, result.WeaponDamage.DiceCount, "Should roll 1 die")
			assert.Equal(t, 4, result.WeaponDamage.DiceSize, "Should be a d4")
		})
	}
}

func TestMonkMartialArts_WithProficiencyBonus(t *testing.T) {
	// This test verifies that proficiency is still added to attack rolls
	// Currently proficiency is handled elsewhere in the attack system

	// Create a level 1 monk with martial arts
	monk := &character.Character{
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{Key: "martial-arts", Name: "Martial Arts"},
		},
		Attributes: map[character.Attribute]*character.AbilityScore{
			character.AttributeStrength: {
				Score: 10, // +0 bonus
				Bonus: 0,
			},
			character.AttributeDexterity: {
				Score: 16, // +3 bonus
				Bonus: 3,
			},
		},
	}

	// Mock dice roller
	mockRoller := mockdice.NewManualMockRoller()
	mockRoller.SetRolls([]int{
		10, // Attack roll
		2,  // Damage roll (1d4)
	})
	monk = monk.WithDiceRoller(mockRoller)

	result, err := monk.improvisedMelee()
	require.NoError(t, err)

	// Attack should be d20 (10) + DEX (3) = 13
	// Note: Proficiency bonus is added elsewhere in the combat system
	assert.Equal(t, 13, result.AttackRoll, "Attack should use DEX for monks")
	assert.Equal(t, 5, result.DamageRoll, "Damage should be 1d4 (2) + DEX (3)")
}

func TestMonkMartialArts_EdgeCases(t *testing.T) {
	t.Run("Monk with equal STR and DEX uses STR", func(t *testing.T) {
		monk := &character.Character{
			Features: []*rulebook.CharacterFeature{
				{Key: "martial-arts", Name: "Martial Arts"},
			},
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeStrength: {
					Score: 14, // +2 bonus
					Bonus: 2,
				},
				character.AttributeDexterity: {
					Score: 14, // +2 bonus
					Bonus: 2,
				},
			},
		}

		mockRoller := mockdice.NewManualMockRoller()
		mockRoller.SetRolls([]int{12, 4})
		monk = monk.WithDiceRoller(mockRoller)

		result, err := monk.improvisedMelee()
		require.NoError(t, err)

		// When equal, STR is used (existing behavior maintained)
		assert.Equal(t, 14, result.AttackRoll, "Should use STR when equal to DEX")
		assert.Equal(t, 6, result.DamageRoll, "Damage uses same ability as attack")
	})

	t.Run("Monk with nil attributes", func(t *testing.T) {
		monk := &character.Character{
			Features: []*rulebook.CharacterFeature{
				{Key: "martial-arts", Name: "Martial Arts"},
			},
			Attributes: nil,
		}

		mockRoller := mockdice.NewManualMockRoller()
		mockRoller.SetRolls([]int{20, 1})
		monk = monk.WithDiceRoller(mockRoller)

		result, err := monk.improvisedMelee()
		require.NoError(t, err)

		assert.Equal(t, 20, result.AttackRoll, "No bonus when attributes are nil")
		assert.Equal(t, 1, result.DamageRoll, "No bonus when attributes are nil")
	})

	t.Run("Character with nil features", func(t *testing.T) {
		char := &character.Character{
			Features: nil,
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeStrength: {
					Score: 10,
					Bonus: 0,
				},
				character.AttributeDexterity: {
					Score: 18,
					Bonus: 4,
				},
			},
		}

		mockRoller := mockdice.NewManualMockRoller()
		mockRoller.SetRolls([]int{15, 3})
		char = char.WithDiceRoller(mockRoller)

		result, err := char.improvisedMelee()
		require.NoError(t, err)

		assert.Equal(t, 15, result.AttackRoll, "Uses STR (0) without martial arts")
		assert.Equal(t, 3, result.DamageRoll, "Uses STR (0) without martial arts")
	})
}
