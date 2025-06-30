package entities_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_Attack_ImprovisedMelee(t *testing.T) {
	tests := []struct {
		name string
		char *character.Character
	}{
		{
			name: "Character with no equipment uses improvised attack",
			char: &character.Character{
				Name:  "Test Fighter",
				Level: 1,
				Attributes: map[shared.Attribute]*character.AbilityScore{
					shared.AttributeStrength: {Score: 14, Bonus: 2},
				},
			},
		},
		{
			name: "Character with nil attributes doesn't crash",
			char: &character.Character{
				Name:       "Broken Character",
				Level:      1,
				Attributes: nil,
			},
		},
		{
			name: "Character with empty equipped slots",
			char: &character.Character{
				Name:          "Empty Handed",
				Level:         1,
				EquippedSlots: make(map[shared.Slot]equipment.Equipment),
				Attributes: map[shared.Attribute]*character.AbilityScore{
					shared.AttributeStrength: {Score: 10, Bonus: 0},
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
	char := &character.Character{
		Name:  "Armed Fighter",
		Level: 3,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
			shared.AttributeDexterity: {Score: 12, Bonus: 1},
		},
		EquippedSlots: map[shared.Slot]equipment.Equipment{
			shared.SlotMainHand: &equipment.Weapon{
				Base: equipment.BasicEquipment{
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
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
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
