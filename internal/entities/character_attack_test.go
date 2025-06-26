package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_Attack_ImprovisedMelee(t *testing.T) {
	tests := []struct {
		name string
		char *entities.Character
	}{
		{
			name: "Character with no equipment uses improvised attack",
			char: &entities.Character{
				Name:  "Test Fighter",
				Level: 1,
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength: {Score: 14, Bonus: 2},
				},
			},
		},
		{
			name: "Character with nil attributes doesn't crash",
			char: &entities.Character{
				Name:       "Broken Character",
				Level:      1,
				Attributes: nil,
			},
		},
		{
			name: "Character with empty equipped slots",
			char: &entities.Character{
				Name:          "Empty Handed",
				Level:         1,
				EquippedSlots: make(map[entities.Slot]entities.Equipment),
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength: {Score: 10, Bonus: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := tt.char.Attack()
			require.NoError(t, err)
			require.Len(t, results, 1)

			result := results[0]
			assert.NotNil(t, result)
			assert.Equal(t, damage.TypeBludgeoning, result.AttackType)

			// Most importantly, these should not be nil (fixing the panic)
			assert.NotNil(t, result.AttackResult, "AttackResult should not be nil")
			assert.NotNil(t, result.DamageResult, "DamageResult should not be nil")

			// Verify the rolls were populated
			assert.Greater(t, result.AttackRoll, 0)
			assert.GreaterOrEqual(t, result.DamageRoll, 1) // At least 1 from 1d4
		})
	}
}

func TestCharacter_Attack_WithWeapon(t *testing.T) {
	char := &entities.Character{
		Name:  "Armed Fighter",
		Level: 3,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:  {Score: 16, Bonus: 3},
			entities.AttributeDexterity: {Score: 12, Bonus: 1},
		},
		EquippedSlots: map[entities.Slot]entities.Equipment{
			entities.SlotMainHand: &entities.Weapon{
				Base: entities.BasicEquipment{
					Key:  "shortsword",
					Name: "Shortsword",
				},
				Damage: &damage.Damage{
					DiceCount:  1,
					DiceSize:   6,
					DamageType: damage.TypeSlashing,
				},
				WeaponRange: "Melee",
			},
		},
		Proficiencies: map[entities.ProficiencyType][]*entities.Proficiency{
			entities.ProficiencyTypeWeapon: {
				{Key: "shortsword", Name: "Shortsword"},
			},
		},
	}

	results, err := char.Attack()
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.NotNil(t, result)
	assert.Equal(t, damage.TypeSlashing, result.AttackType)

	// These should not be nil
	assert.NotNil(t, result.AttackResult, "AttackResult should not be nil")
	assert.NotNil(t, result.DamageResult, "DamageResult should not be nil")
}
